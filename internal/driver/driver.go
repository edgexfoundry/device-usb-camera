// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
// Copyright (C) 2023 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/models"

	"github.com/edgexfoundry/device-sdk-go/v4/pkg/interfaces"
	sdkModels "github.com/edgexfoundry/device-sdk-go/v4/pkg/models"

	"github.com/labstack/echo/v4"
	usbDevice "github.com/vladimirvivien/go4vl/device"
	"github.com/xfrr/goffmpeg/transcoder"
)

var driver *Driver
var once sync.Once

type Driver struct {
	ds                          interfaces.DeviceServiceSDK
	lc                          logger.LoggingClient
	wg                          *sync.WaitGroup
	asyncCh                     chan<- *sdkModels.AsyncValues
	deviceCh                    chan<- []sdkModels.DiscoveredDevice
	activeDevices               map[string]*Device
	rtspHostName                string
	rtspTcpPort                 string
	rtspAuthenticationServerUri string
	mutex                       sync.Mutex
	rtspAuthServer              *echo.Echo
	rtspServerMode              RTSPServerMode
}

// NewProtocolDriver initializes the singleton Driver and returns it to the caller
func NewProtocolDriver() *Driver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

type MultiErr []error

//goland:noinspection GoReceiverNames
func (me MultiErr) Error() string {
	strs := make([]string, len(me))
	for i, s := range me {
		strs[i] = s.Error()
	}

	return strings.Join(strs, "; ")
}

// Initialize performs protocol-specific initialization for the device service.
func (d *Driver) Initialize(sdk interfaces.DeviceServiceSDK) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.lc = sdk.LoggingClient()
	d.asyncCh = sdk.AsyncValuesChannel()
	d.deviceCh = sdk.DiscoveredDeviceChannel()
	d.ds = sdk
	d.activeDevices = make(map[string]*Device)
	d.wg = new(sync.WaitGroup)

	if err := d.ds.AddCustomRoute(common.ApiBase+ApiRefreshDevicePaths, interfaces.Unauthenticated, echo.WrapHandler(http.HandlerFunc(d.RefreshExistingDevicePathsRoute)), http.MethodPost); err != nil {
		return fmt.Errorf("failed to add API route %s, error: %s", ApiRefreshDevicePaths, err.Error())
	}

	var err error
	// if RtspServerMode config parameter is empty, then it should default to
	// "internal" to retain backwards-compatibility
	d.rtspServerMode = RTSPServerMode(strings.ToLower(d.ds.DriverConfigs()[RtspServerMode]))
	if d.rtspServerMode == "" {
		d.rtspServerMode = RTSPServerModeInternal
	} else if d.rtspServerMode == RTSPServerModeNone {
		return nil // nothing left to do
	} else if d.rtspServerMode != RTSPServerModeInternal && d.rtspServerMode != RTSPServerModeExternal {
		return fmt.Errorf("%s value of \"%s\" is invalid. valid options are \"internal\", \"external\", and \"none\"",
			RtspServerMode, d.rtspServerMode)
	}

	rtspAuthenticationServerUri, ok := d.ds.DriverConfigs()[RtspAuthenticationServer]
	if !ok {
		rtspAuthenticationServerUri = DefaultRtspAuthenticationServer
		d.lc.Warnf("service config %s not found. Use the default value: %s", RtspAuthenticationServer, DefaultRtspAuthenticationServer)
	}
	d.lc.Infof("RtspAuthenticationServer: %s", rtspAuthenticationServerUri)
	d.rtspAuthenticationServerUri = rtspAuthenticationServerUri

	if err := d.ds.SecretProvider().RegisterSecretUpdatedCallback(RtspAuthSecretName, d.secretUpdated); err != nil {
		d.lc.Errorf("failed to register secret update callback: %v", err)
	}

	rtspServerHostName, ok := d.ds.DriverConfigs()[RtspServerHostName]
	if !ok {
		rtspServerHostName = DefaultRtspServerHostName
		d.lc.Warnf("service config %s not found. Use the default value: %s", RtspServerHostName, DefaultRtspServerHostName)
	}
	d.lc.Infof("RTSP server hostname: %s", rtspServerHostName)
	d.rtspHostName = rtspServerHostName

	rtspPort, ok := d.ds.DriverConfigs()[RtspTcpPort]
	if !ok {
		rtspPort = DefaultRtspTcpPort
		d.lc.Warnf("service config %s not found. Use the default value: %s", RtspTcpPort, DefaultRtspTcpPort)
	}
	d.lc.Infof("RTSP TCP port: %s", rtspPort)
	d.rtspTcpPort = rtspPort

	if d.rtspServerMode != RTSPServerModeInternal {
		return nil // nothing left to do
	}

	// check to see if mediamtx file/binary exists
	rtspExecutable := d.ds.DriverConfigs()[RtspServerExe]
	if rtspExecutable == "" {
		// to ensure backwards compatibility
		rtspExecutable = RtspServerExeDefault
	}
	_, err = os.Stat(rtspExecutable)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s file cannot be found: %s", RtspServerExe, err.Error())
		} else if os.IsPermission(err) {
			return fmt.Errorf("permission denied for %s: %s", RtspServerExe, err.Error())
		} else {
			return fmt.Errorf("unknown error occurred while checking for rtsp-server binary file %s: %s", rtspExecutable, err.Error())
		}
	}

	rtspProc := exec.Command(rtspExecutable)
	rtspProc.Stdout = os.Stdout
	rtspProc.Stderr = os.Stderr
	err = rtspProc.Start()
	if err != nil {
		return fmt.Errorf("unable to start %s process: %s", rtspExecutable, err.Error())
	}

	return nil
}

func (d *Driver) Start() error {
	d.lc.Info("Initializing cameras...")
	for _, dev := range d.ds.Devices() {
		d.lc.Infof("initialize device %s", dev.Name)
		activeDevice, edgexErr := d.newDevice(dev.Name, dev.Protocols)
		if edgexErr != nil {
			d.lc.Error(edgexErr.Error())
			continue
		}
		d.activeDevices[dev.Name] = activeDevice
	}

	// Make sure the paths of existing devices are up-to-date.
	go d.RefreshAllDevicePaths()

	if d.rtspServerMode == RTSPServerModeNone {
		d.lc.Info("RTSP server is disabled")
		return nil
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.StartRTSPCredentialServer()
	}()

	for _, dev := range d.activeDevices {
		if dev.autoStreaming {
			dev.streamingStatus.TranscoderInputPath = dev.paths[0]
			edgexErr := d.startStreaming(dev)
			if edgexErr != nil {
				d.lc.Errorf("failed to start video streaming for device %s, error: %s", dev.name, edgexErr)
			}

		}
	}

	return nil
}

func (d *Driver) StartRTSPCredentialServer() {
	d.lc.Infof("Starting rtsp authentication server on %s", d.rtspAuthenticationServerUri)

	e := echo.New()
	e.HideBanner = true
	e.Server.ReadHeaderTimeout = 5 * time.Second // G112: A configured ReadHeaderTimeout in the http.Server averts a potential Slowloris Attack
	e.Router().Add(http.MethodPost, "/rtspauth", d.RTSPCredentialsHandler)
	d.rtspAuthServer = e

	err := e.Start(d.rtspAuthenticationServerUri)
	if err != nil && err != http.ErrServerClosed {
		d.lc.Errorf("RTSP Auth Web server failed: %v", err)
		d.rtspAuthServer = nil
	}
}

// RTSPCredentialsHandler is the http handler for handling rtsp authentication requests. It expects the
// request to follow the format defined by https://github.com/aler9/mediamtx#authentication
func (d *Driver) RTSPCredentialsHandler(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		d.lc.Errorf("rtsp authentication: could not read body: %s", err)
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	var rtspAuthRequest RTSPAuthRequest
	err = json.Unmarshal(body, &rtspAuthRequest)
	if err != nil {
		d.lc.Errorf("rtsp authentication: could not unmarshal body into json: %s", err)
		c.Response().WriteHeader(http.StatusBadRequest)
		return nil
	}

	if rtspAuthRequest.User == "" || rtspAuthRequest.Password == "" {
		d.lc.Debug("rtsp authentication: username or password is empty")
		// From https://github.com/aler9/mediamtx#authentication README:
		// Please be aware that it's perfectly normal for the authentication server to receive requests with
		// empty users and passwords. This happens because a RTSP client doesn't provide credentials until it
		// is asked to. In order to receive the credentials, the authentication server must reply with status code 401,
		// then the client will send credentials.
		c.Response().WriteHeader(http.StatusUnauthorized)
		return nil
	}
	credential, edgexErr := d.tryGetCredentials(RtspAuthSecretName)
	if edgexErr != nil {
		d.lc.Warnf("Failed to retrieve credentials for rtsp authentication from the secret store. Have you stored credentials yet for secretName %s?", RtspAuthSecretName)
		c.Response().WriteHeader(http.StatusInternalServerError)
		if _, err = c.Response().Write([]byte("RTSP Authentication has not been fully configured!")); err != nil {
			d.lc.Errorf("Error writing message: %v", err.Error())
		}
		return nil
	}

	if credential.Username != rtspAuthRequest.User ||
		credential.Password != rtspAuthRequest.Password {
		d.lc.Warn("rtsp authentication: user or password do not match")
		c.Response().WriteHeader(http.StatusUnauthorized)
		return nil
	}

	d.lc.Debug("rtsp authentication: passwords match")
	return nil
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest) ([]*sdkModels.CommandValue, error) {
	d.lc.Debugf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols,
		reqs[0].DeviceResourceName, reqs[0].Attributes)
	var responses = make([]*sdkModels.CommandValue, len(reqs))

	device, edgexErr := d.getDevice(deviceName)
	if edgexErr != nil {
		return responses, edgexErr
	}

	for i, req := range reqs {
		command, ok := req.Attributes[GetFunction]
		if !ok {
			return responses, errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("command for USB camera resource %s is not specified, please check device profile",
					req.DeviceResourceName), nil)
		}
		cv, err := d.ExecuteReadCommands(device, req, command)
		if err != nil {
			// flush query parameter for remaining reqs
			for _, req := range reqs {
				if _, ok := req.Attributes[UrlRawQuery]; ok {
					req.Attributes[UrlRawQuery] = ""
				}
			}
			return responses, err
		}
		// flush query parameter
		if _, ok := req.Attributes[UrlRawQuery]; ok {
			req.Attributes[UrlRawQuery] = ""
		}
		responses[i] = cv
	}

	return responses, nil
}

func (d *Driver) ExecuteReadCommands(device *Device, req sdkModels.CommandRequest, command interface{}) (*sdkModels.CommandValue, error) {
	var cv *sdkModels.CommandValue
	var data interface{}
	errorWrapper := EdgeXErrorWrapper{}

	queryParams, edgexErr := getQueryParameters(req)
	if edgexErr != nil {
		return cv, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	videoPath, err := d.getPathName(device, queryParams)
	if err != nil {
		return nil, err
	}

	cameraDevice, err := usbDevice.Open(videoPath)
	if err != nil {
		return cv, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to open the underlying device at specified path %s", videoPath), err)
	}
	defer cameraDevice.Close()

	switch command := fmt.Sprintf("%v", command); command {
	case MetadataDeviceCapability:
		data, err = getCapability(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case MetadataCurrentVideoInput:
		data, err = cameraDevice.GetVideoInputIndex()
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, data)
	case MetadataCameraStatus:
		index := queryParams.Get(InputIndex)
		if len(index) == 0 {
			return nil, fmt.Errorf("mandatory query parameter %s not found", InputIndex)
		}
		data, err = getInputStatus(cameraDevice, index)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, data)
	case MetadataImageFormats:
		data, err = getImageFormats(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case MetadataFrameRateFormats:
		data, err = getSupportedFrameRateFormats(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case VideoGetFrameRate:
		data, err = GetFrameRate(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case VideoGetPixelFormat:
		data, err = device.GetPixelFormat(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case MetadataDataFormat:
		data, err = getDataFormat(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case MetadataCroppingAbility:
		data, err = getCropInfo(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case MetadataStreamingParameters:
		data, err = getStreamingParameters(cameraDevice)
		if err != nil {
			return nil, errorWrapper.CommandError(command, err)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
	case VideoStreamUri:
		if d.rtspServerMode == RTSPServerModeNone {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
				"rtsp server is not enabled, cannot get stream URI for device %s", device.name), nil)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, req.Type, device.rtspUri)
	case VideoStreamingStatus:
		if d.rtspServerMode == RTSPServerModeNone {
			return nil, errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
				"rtsp server is not enabled, cannot get streaming status for device %s", device.name), nil)
		}
		cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, device.streamingStatus)
	default:
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported command %s", command), nil)
	}
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError, "failed to create CommandValue", err)
	}
	return cv, nil
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest, params []*sdkModels.CommandValue) error {
	device, edgexErr := d.getDevice(deviceName)
	if edgexErr != nil {
		return edgexErr
	}

	for i, req := range reqs {
		command, ok := req.Attributes[SetFunction]
		if !ok {
			return errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("command for USB camera resource %s is not specified, please check device profile",
					req.DeviceResourceName), nil)
		}

		err := d.ExecuteWriteCommands(device, req, params[i], command)
		if err != nil {
			// flush query parameter for remaining reqs
			for _, req := range reqs {
				if _, ok := req.Attributes[UrlRawQuery]; ok {
					req.Attributes[UrlRawQuery] = ""
				}
			}
			return err
		}
		// flush query parameter
		if _, ok := req.Attributes[UrlRawQuery]; ok {
			req.Attributes[UrlRawQuery] = ""
		}
	}
	return nil
}

func (d *Driver) ExecuteWriteCommands(device *Device, req sdkModels.CommandRequest, param *sdkModels.CommandValue, command interface{}) error {
	queryParams, edgexErr := getQueryParameters(req)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	videoPath, err := d.getPathName(device, queryParams)
	if err != nil {
		return err
	}

	cameraDevice, err := usbDevice.Open(videoPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to open the underlying device at specified path %s", videoPath), err)
	}
	defer cameraDevice.Close()

	switch command {
	case VideoStartStreaming:
		err = device.updateTranscoderInputPath(videoPath)
		if err != nil {
			return err
		}
		device.streamingStatus.TranscoderInputPath = videoPath
		options, edgexErr := param.ObjectValue()
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		edgexErr = setupFFmpegOptions(device, options, req.Attributes)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		edgexErr = d.startStreaming(device)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	case VideoStopStreaming:
		if d.rtspServerMode == RTSPServerModeNone {
			return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
				"rtsp server is not enabled, cannot stop streaming for device %s", device.name), nil)
		}
		device.StopStreaming()
	case VideoSetFrameRate:
		frameRateParam, edgexErr := param.ObjectValue()
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		var frameRateDenominator uint64
		frameRateValueDenominator, ok := frameRateParam.(map[string]interface{})[FrameRateValueDenominator]
		if !ok {
			frameRateDenominator = 1
		} else {
			frameRateDenominator, err = strconv.ParseUint(frameRateValueDenominator.(string), 0, 32)
			if err != nil {
				d.lc.Errorf("Could not parse numerator %d to uint32", frameRateValueDenominator)
				return err
			}
		}

		frameRateValueNumerator, ok := frameRateParam.(map[string]interface{})[FrameRateValueNumerator]
		if !ok {
			return errors.NewCommonEdgeXWrapper(nil)
		}
		frameRateNumerator, err := strconv.ParseUint(frameRateValueNumerator.(string), 0, 32)
		if err != nil {
			d.lc.Errorf("Could not parse denominator %d to uint32", frameRateNumerator)
			return err
		}
		fps, err := device.SetFrameRate(cameraDevice, uint32(frameRateNumerator), uint32(frameRateDenominator))
		if err != nil {
			d.lc.Errorf("Could not set the FPS to %f for device %s due to error: %s", fps, device.name, err)
			return err
		}
		d.lc.Infof("Device frame rate set to %s", fps)
	case VideoSetPixelFormat:
		params, edgexErr := param.ObjectValue()
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		edgexErr = device.SetPixelFormat(cameraDevice, params)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
		d.lc.Infof("Pixel format set for the device %s", device.name)
	default:
		return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported command %s", command), nil)
	}
	return nil
}

func (d *Driver) Stop(force bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// The wait group is used here as well as in the startStreaming functions.
	// The call to Wait() waits for StopStreaming to return and startStreaming to end.
	defer d.wg.Wait()

	if d.rtspServerMode == RTSPServerModeNone {
		return nil
	}

	if d.rtspAuthServer != nil {
		err := d.rtspAuthServer.Shutdown(context.Background())
		if err != nil {
			d.lc.Errorf("Error occurred while shutting down the rtsp auth server: %v", err)
		}
	}

	d.wg.Add(len(d.activeDevices))

	for _, device := range d.activeDevices {
		go func(device *Device) {
			device.StopStreaming()
			d.wg.Done()
		}(device)
	}
	return nil

}

func (d *Driver) getPathName(device *Device, queryParams url.Values) (string, error) {
	var videoPath string
	pathIndex := queryParams.Get(PathIndex)
	streamFormat := queryParams.Get(StreamFormat)
	if pathIndex == "" && streamFormat == "" { // most common case
		videoPath = device.paths[0]
	} else if pathIndex != "" && streamFormat != "" { // both cannot be provided
		return "", errors.NewCommonEdgeX(errors.KindIOError, "Cannot provide both PathIndex and StreamFormat query parameters to command", nil)
	} else if pathIndex != "" { // use path index video path value
		pathIndexConv, err := strconv.Atoi(pathIndex)
		if err != nil {
			return "", err
		}
		if pathIndexConv >= len(device.paths) {
			return "", errors.NewCommonEdgeX(errors.KindIOError,
				fmt.Sprintf("Video streaming path does not exist for the device %v at PathIndex %d", device.name, pathIndexConv), nil)
		}
		videoPath = device.paths[pathIndexConv]
	} else if streamFormat != "" { // use stream format video path value
		if streamFormat != RGB && streamFormat != Greyscale && streamFormat != Depth {
			return "", errors.NewCommonEdgeX(errors.KindIOError, "Invalid stream format. Valid options are 'RGB', 'Greyscale', or 'Depth.'", nil)
		}
		for _, p := range device.paths {
			if d.pathMatchesStreamFormat(p, streamFormat) {
				return p, nil
			}
		}
		return "", errors.NewCommonEdgeX(errors.KindIOError, fmt.Sprintf("Invalid stream format for device %s.", device.name), nil)
	}
	return videoPath, nil
}

func (d *Driver) pathMatchesStreamFormat(path, streamFormat string) bool {
	formatDevice, err := usbDevice.Open(path)
	if err != nil {
		d.lc.Errorf("Cannot open device at path %s", path)
		return false
	}
	defer formatDevice.Close()
	formatDescriptions, err := formatDevice.GetFormatDescriptions()
	if err != nil {
		d.lc.Errorf("Cannot get formatDescriptions at path %s", path)
		return false
	}
	currentType := StreamFormatTypeMap[formatDescriptions[0].PixelFormat]
	return currentType == streamFormat
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties,
	adminState models.AdminState) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_, err := d.addDeviceInternal(deviceName, protocols)
	return err
}

// addDeviceInternal attempts to add a device to the device service's active devices
func (d *Driver) addDeviceInternal(deviceName string, protocols map[string]models.ProtocolProperties) (*Device, error) {
	paths, err := d.getPaths(protocols)
	if err != nil {
		return nil, err
	}

	_, sn, edgexErr := getUSBDeviceIdInfo(paths[0])
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("could not find the serial number of the device %s", deviceName), edgexErr)
	}
	for _, ad := range d.activeDevices {
		if ad.serialNumber == sn {
			return nil, errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("the serial number %s conflicts with existing device %s", sn, ad.name), nil)
		}
	}
	activeDevice, edgexErr := d.newDevice(deviceName, protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	d.activeDevices[deviceName] = activeDevice
	d.lc.Debugf("a new Device is added: %s", deviceName)
	if activeDevice.autoStreaming {
		activeDevice.streamingStatus.TranscoderInputPath = paths[0]
		edgexErr = d.startStreaming(activeDevice)
		if edgexErr != nil {
			return nil, errors.NewCommonEdgeXWrapper(edgexErr)
		}

	}
	return activeDevice, nil
}

// UpdateDevice is a callback function that is invoked
// when a Device associated with this Device Service is updated
func (d *Driver) UpdateDevice(deviceName string, protocols map[string]models.ProtocolProperties,
	adminState models.AdminState) error {
	if edgexErr := d.RemoveDevice(deviceName, protocols); edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	if edgexErr := d.AddDevice(deviceName, protocols, adminState); edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	d.lc.Debugf("Device %s is updated", deviceName)
	return nil
}

// RemoveDevice is a callback function that is invoked
// when a Device associated with this Device Service is removed
func (d *Driver) RemoveDevice(deviceName string, protocols map[string]models.ProtocolProperties) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if device, ok := d.activeDevices[deviceName]; ok {
		if d.rtspServerMode != RTSPServerModeNone {
			device.StopStreaming()
		}
		delete(d.activeDevices, deviceName)
		d.lc.Debugf("Device %s is removed", deviceName)
	}
	return nil
}

// RefreshAllDevicePaths runs RefreshDevicePaths for every connected device
func (d *Driver) RefreshAllDevicePaths() {
	d.lc.Debug("Refreshing existing device paths")
	for _, cd := range d.ds.Devices() {
		d.RefreshDevicePaths(cd)
	}
}

// Deprecated: 3.0 version of the service supports Path which is a string instead of Paths which is []string which leads to breaking change.
// updatePathToPaths takes care of this breaking change.
func (d *Driver) updatePathToPaths(cd models.Device) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if value, ok := cd.Protocols[UsbProtocol][Path]; ok {
		if value != nil {
			cd.Protocols[UsbProtocol][Paths] = []string{value.(string)}
			delete(cd.Protocols[UsbProtocol], "Path")

			if err := d.ds.PatchDevice(dtos.UpdateDevice{
				Name:      &cd.Name,
				Protocols: dtos.FromProtocolModelsToDTOs(cd.Protocols),
			}); err != nil {
				d.lc.Errorf("failed to update the device %s, error: %s", cd.Name, err.Error())
			}
		}
	}
}

// RefreshDevicePaths checks whether the existing device matches the connected device.
// If there is a mismatch between them, scan all paths to find the matching device
// and update the existing device with the correct path.
func (d *Driver) RefreshDevicePaths(cd models.Device) {
	d.updatePathToPaths(cd)

	paths, err := d.getPaths(cd.Protocols)
	if err != nil {
		d.lc.Errorf("Failed to get paths for device %s", cd.Name)
	}
	for _, fdPath := range paths {
		_, sn, err := getUSBDeviceIdInfo(fdPath)
		if err != nil {
			d.lc.Errorf("failed to get the serial number of device %s, error: %s", cd.Name, err.Error())
		}
		// If the serial number is different, it means that the path of the device has changed.
		if sn != cd.Protocols[UsbProtocol][SerialNumber] || !d.isVideoCaptureDevice(fdPath) {
			// Delete the paths and start fresh
			cd.Protocols[UsbProtocol][Paths] = nil
			go d.updateDevicePaths(cd)
			break
		}
	}
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (d *Driver) Discover() error {
	d.lc.Info("Discovery is triggered")
	devices := make(map[string]sdkModels.DiscoveredDevice)

	// Convert the slice of cached devices to map in order to improve the performance in the subsequent for loop.
	currentDevices := d.cachedDeviceMap()

	allDevices, _ := usbDevice.GetAllDevicePaths()
	for _, fdPath := range allDevices {
		if ok := d.isVideoCaptureDevice(fdPath); ok {
			cn, sn, err := getUSBDeviceIdInfo(fdPath)
			if err != nil {
				d.lc.Errorf("failed to get device serial number, error: %s", err.Error())
				continue
			}
			// Update existing device if it's path has changed
			if _, ok := currentDevices[cn+sn]; ok {
				d.RefreshDevicePaths(currentDevices[cn+sn])
				continue
			}
			if _, found := devices[cn+sn]; !found {
				discovered := sdkModels.DiscoveredDevice{
					Name: buildDeviceName(cn, sn),
					Protocols: map[string]models.ProtocolProperties{
						UsbProtocol: {
							Paths:        []string{fdPath},
							SerialNumber: sn,
							CardName:     cn,
						},
					},
					Description: fmt.Sprintf("USB camera %s", cn),
					Labels:      []string{"auto-discovery", cn},
				}
				devices[cn+sn] = discovered
				d.lc.Infof("discovered device: %s", discovered.Name)
			} else {
				devices[cn+sn].Protocols[UsbProtocol][Paths] = append(devices[cn+sn].Protocols[UsbProtocol][Paths].([]string), fdPath)
			}
		}
	}
	var discoveredDevices []sdkModels.DiscoveredDevice
	for _, device := range devices {
		discoveredDevices = append(discoveredDevices, device)
	}
	d.deviceCh <- discoveredDevices
	return nil
}

func (d *Driver) getProtocolProperty(protocols map[string]models.ProtocolProperties, protocol, key string) (string, errors.EdgeX) {
	if _, ok := protocols[protocol]; !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("%s protocol configuration not found. Please check device configuration", protocol), nil)
	}
	value, ok := protocols[protocol][key]
	if !ok {
		return "", errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("property %s of protocol %s is missing. Please check device configuration",
				key, protocol), nil)
	}
	return value.(string), nil
}

func (d *Driver) getPaths(protocols map[string]models.ProtocolProperties) ([]string, errors.EdgeX) {
	if _, ok := protocols[UsbProtocol]; !ok {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("%s protocol configuration not found. Please check device configuration", UsbProtocol), nil)
	}
	value, ok := protocols[UsbProtocol][Paths]
	if !ok {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("property %s of protocol %s is missing. Please check device configuration",
				Paths, UsbProtocol), nil)
	}
	if value != nil {
		// Depending on where the function is called from, protocols could contain
		// a []string or []interface{} filled with strings
		if _, ok := value.([]interface{}); ok {
			s := make([]string, len(value.([]interface{})))
			for index, v := range value.([]interface{}) {
				s[index] = v.(string)
			}
			return s, nil
		}
		return value.([]string), nil
	} else {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("property %s of protocol %s is missing. Please check device configuration",
				Paths, UsbProtocol), nil)
	}
}

func (d *Driver) newDevice(name string, protocols map[string]models.ProtocolProperties) (*Device, errors.EdgeX) {
	device, err := d.ds.GetDeviceByName(name)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("device %s not found in core metadata", name), err)
	}
	profile, err := d.ds.GetProfileByName(device.ProfileName)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("profile %s not found in core metadata", name), err)
	}

	paths, edgexErr := d.getPaths(protocols)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	fdPath := paths[0]

	rtspUri := &url.URL{
		Scheme: RtspUriScheme,
		Host:   fmt.Sprintf("%s:%s", d.rtspHostName, d.rtspTcpPort),
	}
	rtspUri.Path = path.Join(Stream, name)

	// Create new instance of transcoder
	trans := new(transcoder.Transcoder)
	// Initialize transcoder passing the input path and output path, along with the credentials
	err = trans.Initialize(fdPath, d.getAuthenticatedRTSPUri(name))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to initialize transcoder for device %s", name), err)
	}
	trans.MediaFile().SetOutputFormat(RtspUriScheme)

	autoStreaming := false
	autoStreamingStr, edgexErr := d.getProtocolProperty(protocols, UsbProtocol, AutoStreaming)
	if edgexErr != nil {
		d.lc.Warnf("Protocol property %s not found. Use default value: %v", AutoStreaming, autoStreaming)
	} else {
		autoStreaming, err = strconv.ParseBool(autoStreamingStr)
		if err != nil {
			d.lc.Errorf("invalid input value %s for protocol property %s. Use default value: %v", autoStreamingStr,
				AutoStreaming, autoStreaming)
		}
	}

	var streamingStatusResourceName string
	for _, r := range profile.DeviceResources {
		command, ok := r.Attributes[GetFunction]
		if ok && command == VideoStreamingStatus {
			streamingStatusResourceName = r.Name
			break
		}
	}
	if len(streamingStatusResourceName) == 0 {
		d.lc.Warnf("there is no device resource representing StreamingStatus of the device %s, so the StreamingStatus won't be published automatically", name)
	}

	cameraDevice, err := usbDevice.Open(fdPath)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to open the underlying device at specified path %s", fdPath), err)
	}
	defer cameraDevice.Close()

	cn, sn, err := getUSBDeviceIdInfo(fdPath)
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("could not find the serial number of the device on the specified path: %s", fdPath), err)
	}

	// if the user provided a serial number or card name, but it does not match the device's serial number or card name,
	// then return an error, as this may not be the correct device
	psn, psnOK := protocols[UsbProtocol][SerialNumber].(string)
	if psnOK && psn != "" && psn != sn {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("wrong device serial number, expected %s=%s, actual %s=%s", SerialNumber, psn, SerialNumber, sn), nil)
	}
	pcn, pcnOK := protocols[UsbProtocol][CardName].(string)
	if pcnOK && pcn != "" && pcn != cn {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("wrong device card name, expected %s=%s, actual %s=%s", CardName, pcn, CardName, cn), nil)
	}

	shouldUpdate := false
	if !psnOK || psn == "" { // pre-defined devices may not include serial number information
		device.Protocols[UsbProtocol][SerialNumber] = sn
		shouldUpdate = true
	}
	if !pcnOK || pcn == "" { // pre-defined devices may not include card name information
		device.Protocols[UsbProtocol][CardName] = cn
		shouldUpdate = true
	}

	if shouldUpdate {
		if err := d.ds.PatchDevice(dtos.UpdateDevice{
			Name:      &device.Name,
			Protocols: dtos.FromProtocolModelsToDTOs(device.Protocols),
		}); err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("failed to update the device %s to add serial number and/or card name information", device.Name), err)
		}
	}

	return &Device{
		lc:                          d.lc,
		name:                        name,
		paths:                       paths,
		serialNumber:                sn,
		rtspUri:                     rtspUri.String(),
		transcoder:                  trans,
		autoStreaming:               autoStreaming,
		streamingStatusResourceName: streamingStatusResourceName,
	}, nil
}

func (d *Driver) getAuthenticatedRTSPUri(name string) string {
	rtspAuthenticatedUri := &url.URL{
		Scheme: RtspUriScheme,
		Host:   fmt.Sprintf("%s:%s", d.rtspHostName, d.rtspTcpPort),
	}
	rtspAuthenticatedUri.Path = path.Join(Stream, name)
	credential, edgexErr := d.tryGetCredentials(RtspAuthSecretName)
	if edgexErr != nil {
		d.lc.Warnf("Failed to retrieve credentials for rtsp authentication from the secret store. Have you stored credentials yet for secretName %s?", RtspAuthSecretName)
	} else {
		rtspAuthenticatedUri.User = url.UserPassword(credential.Username, credential.Password)
	}
	return rtspAuthenticatedUri.String()
}

// getDevice gets an active device by name, which is managed by device service.
func (d *Driver) getDevice(name string) (*Device, errors.EdgeX) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	device, ok := d.activeDevices[name]
	if ok {
		return device, nil
	} else {
		// lookup device to find protocol properties
		edgeXDevice, err := d.ds.GetDeviceByName(name)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("device %s not found in core metadata", name), err)
		}
		// try to add device
		device, err = d.addDeviceInternal(name, edgeXDevice.Protocols)
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindCommunicationError, fmt.Sprintf("device service unable to communicate with device %s", name), err)
		}
		return device, nil
	}
}

func (d *Driver) startStreaming(device *Device) errors.EdgeX {
	// check to see if rtsp server is enabled
	if d.rtspServerMode == RTSPServerModeNone {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
			"rtsp server is not enabled, cannot start streaming for device %s", device.name), nil)
	}

	progressChan, errChan, err := device.StartStreaming()
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
			"failed to start video streaming for device %s", device.name), err)
	}

	defer func() {
		// before we return, publish the current streaming status right away.
		// do this even on error, as this will contain the error message as well.
		go d.publishStreamingStatus(device)
	}()

	waitForFinishAndPublish := func() {
		d.lc.Debugf("Waiting for ffmpeg errChan to be done")
		<-errChan
		d.lc.Debugf("Done waiting for ffmpeg errChan to be done")
		d.publishStreamingStatus(device)
	}

	// wait a little bit before returning to see if there are any errors on startup
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// this should rarely happen, as ffmpeg should print progress on the first frame. If it does happen,
			// then either progress has been disabled or the process could be having issues.
			d.lc.Warnf("Video streaming for device %s has started but has not sent progress messages yet.", device.name)
			go waitForFinishAndPublish() // track process in the background
			return nil
		case startErr, ok := <-errChan:
			if startErr == nil || !ok {
				d.lc.Warnf("Video streaming for device %s seems to have stopped already.", device.name)
				return nil
			}
			return errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("the video streaming process for device %s has stopped", device.name), startErr)
		case _, ok := <-progressChan:
			if !ok {
				continue // channel was closed, so something else must be the problem
			}
			// if we got a progress message, that means that the transcoding is successful
			d.lc.Infof("Video streaming for device %s has started without error", device.name)
			go waitForFinishAndPublish() // track process in the background
			return nil
		}
	}
}

// publishStreamingStatus asynchronously sends an event of StreamingStatus to the Core Metadata service.
func (d *Driver) publishStreamingStatus(device *Device) {
	if len(device.streamingStatusResourceName) == 0 {
		return
	}
	cv, err := sdkModels.NewCommandValue(device.streamingStatusResourceName, common.ValueTypeObject, device.streamingStatus)
	if err != nil {
		d.lc.Error(err.Error())
		return
	}
	asyncValues := &sdkModels.AsyncValues{
		DeviceName:    device.name,
		CommandValues: []*sdkModels.CommandValue{cv},
	}
	d.asyncCh <- asyncValues
}

// cachedDeviceMap return a map of cached devices. Key is a string consists of card name and serial number.
func (d *Driver) cachedDeviceMap() map[string]models.Device {
	cds := d.ds.Devices()
	cdm := make(map[string]models.Device, len(cds))
	for _, cd := range cds {
		cn, cnOK := cd.Protocols[UsbProtocol][CardName].(string)
		sn, snOK := cd.Protocols[UsbProtocol][SerialNumber].(string)
		if len(cn) > 0 && cnOK && len(sn) > 0 && snOK {
			cdm[cn+sn] = cd
		}
	}
	return cdm
}

func (d *Driver) isVideoCaptureDevice(path string) bool {
	cameraDevice, err := usbDevice.Open(path)
	if err != nil {
		d.lc.Debugf("there is no USB camera at specified path %s, error: %s", path, err.Error())
		return false
	}
	defer cameraDevice.Close()
	c := cameraDevice.Capability()
	return isVideoCaptureSupported(c) && isStreamingSupported(c)
}

func (d *Driver) updateDevicePaths(device models.Device) {
	oldPaths := device.Protocols[UsbProtocol][Paths]
	var init []string
	device.Protocols[UsbProtocol][Paths] = init
	allDevices, _ := usbDevice.GetAllDevicePaths()
	for _, fdPath := range allDevices {
		if ok := d.isVideoCaptureDevice(fdPath); ok {
			cn, sn, err := getUSBDeviceIdInfo(fdPath)
			if err != nil {
				d.lc.Errorf("failed to get device serial number, path=%s, error: %s", fdPath, err.Error())
				continue
			}
			if cn == device.Protocols[UsbProtocol][CardName] && sn == device.Protocols[UsbProtocol][SerialNumber] {
				device.Protocols[UsbProtocol][Paths] = append(device.Protocols[UsbProtocol][Paths].([]string), fdPath)
			}
		}
	}

	if !slicesAreEqual(d.lc, device.Protocols[UsbProtocol][Paths], oldPaths) {
		if err := d.ds.PatchDevice(dtos.UpdateDevice{
			Name:      &device.Name,
			Protocols: dtos.FromProtocolModelsToDTOs(device.Protocols),
		}); err != nil {
			d.lc.Errorf("failed to update paths for the device %s", device.Name)
		}
	}
}

func getQueryParameters(req sdkModels.CommandRequest) (url.Values, errors.EdgeX) {
	urlRawQuery := req.Attributes[UrlRawQuery]
	queryParams, err := url.ParseQuery(fmt.Sprintf("%v", urlRawQuery))
	if err != nil {
		return nil, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("invalid query parameters: %s", urlRawQuery), err)
	}
	return queryParams, nil
}

// getUSBDeviceIdInfo returns the serial number and the card name of the device on the specified path
func getUSBDeviceIdInfo(path string) (cardName string, serialNumber string, err error) {
	cmd := exec.Command("udevadm", "info", "--query=property", path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to run command: %s: %s", cmd.String(), output), err)
	}
	props := strings.Split(string(output), "\n")
	m := make(map[string]string, len(props))
	for _, prop := range props {
		kvp := strings.Split(prop, "=")
		if len(kvp) == 2 {
			m[kvp[0]] = kvp[1]
		}
	}
	cardName = m[UdevV4lProduct]
	if len(cardName) == 0 {
		return "", "", errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("could not find the card name of the device on the specified path %s", path), nil)
	}
	if len(m[UdevSerialShort]) > 0 {
		serialNumber = m[UdevSerialShort]
	} else {
		serialNumber = m[UdevSerial]
	}
	if len(serialNumber) == 0 {
		return "", "", errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("could not find the serial number of the device on the specified path %s", path), nil)
	}
	return
}

func (d *Driver) ValidateDevice(device models.Device) error {
	_, err := d.getPaths(device.Protocols)
	if err != nil {
		return errors.NewCommonEdgeXWrapper(err)
	}
	return nil
}
