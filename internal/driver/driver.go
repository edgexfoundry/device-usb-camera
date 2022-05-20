// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"

	sdkModels "github.com/edgexfoundry/device-sdk-go/v2/pkg/models"
	"github.com/edgexfoundry/device-sdk-go/v2/pkg/service"

	usbdevice "github.com/vladimirvivien/go4vl/v4l2/device"
	"github.com/xfrr/goffmpeg/transcoder"
)

var driver *Driver
var once sync.Once

type Driver struct {
	ds            *service.DeviceService
	lc            logger.LoggingClient
	wg            *sync.WaitGroup
	asyncCh       chan<- *sdkModels.AsyncValues
	deviceCh      chan<- []sdkModels.DiscoveredDevice
	activeDevices map[string]*Device
	rtspHostName  string
	rtspTcpPort   string
	mutex         sync.Mutex
}

// NewProtocolDriver initializes the singleton Driver and returns it to the caller
func NewProtocolDriver() *Driver {
	once.Do(func() {
		driver = new(Driver)
	})
	return driver
}

// Initialize performs protocol-specific initialization for the device service.
func (d *Driver) Initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModels.AsyncValues,
	deviceCh chan<- []sdkModels.DiscoveredDevice) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.lc = lc
	d.asyncCh = asyncCh
	d.deviceCh = deviceCh
	d.ds = service.RunningService()
	d.activeDevices = make(map[string]*Device)
	d.wg = new(sync.WaitGroup)

	h := NewHttpHandler(d)
	if err := d.ds.AddRoute(common.ApiBase+ApiRefreshDevicePaths, h.RefreshExistingDevicePaths, http.MethodPost); err != nil {
		return fmt.Errorf("failed to add API route %s, error: %s", ApiRefreshDevicePaths, err.Error())
	}

	rtspServerHostName, ok := service.DriverConfigs()[RtspServerHostName]
	if !ok {
		rtspServerHostName = DefaultRtspServerHostName
		d.lc.Warnf("service config %s not found. Use the default value: %s", RtspServerHostName, DefaultRtspServerHostName)
	}
	d.lc.Infof("RTSP server hostname: %s", rtspServerHostName)
	d.rtspHostName = rtspServerHostName

	rtspPort, ok := service.DriverConfigs()[RtspTcpPort]
	if !ok {
		rtspPort = DefaultRtspTcpPort
		d.lc.Warnf("service config %s not found. Use the default value: %s", RtspTcpPort, DefaultRtspTcpPort)
	}
	d.lc.Infof("RTSP TCP port: %s", rtspPort)
	d.rtspTcpPort = rtspPort

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

	for _, dev := range d.activeDevices {
		if dev.autoStreaming {
			edgexErr := d.startStreaming(dev)
			if edgexErr != nil {
				d.lc.Errorf("failed to start video streaming for device %s, error: %s", dev.name, edgexErr)
			}
		}
	}

	// Make sure the paths of existing devices are up to date.
	go d.RefreshExistingDevicePaths()

	return nil
}

func (d *Driver) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest) ([]*sdkModels.CommandValue, error) {
	d.lc.Debugf("Driver.HandleReadCommands: protocols: %v resource: %v attributes: %v", protocols,
		reqs[0].DeviceResourceName, reqs[0].Attributes)
	var err error
	var responses = make([]*sdkModels.CommandValue, len(reqs))

	device, edgexErr := d.getDevice(deviceName)
	if edgexErr != nil {
		return responses, errors.NewCommonEdgeXWrapper(edgexErr)
	}
	cameraDevice, err := usbdevice.Open(device.path)
	if err != nil {
		return responses, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to open the underlying device at specified path %s", device.path), err)
	}
	defer cameraDevice.Close()

	var cv *sdkModels.CommandValue
	var data interface{}
	errorWrapper := EdgeXErrorWrapper{}
	for i, req := range reqs {
		command, ok := req.Attributes[Command]
		if !ok {
			return responses, errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("command for USB camera resource %s is not specified, please check device profile",
					req.DeviceResourceName), nil)
		}

		switch command := fmt.Sprintf("%v", command); command {
		case MetadataDeviceCapability:
			data, err = getCapability(cameraDevice)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
		case MetadataCurrentVideoInput:
			data, err = cameraDevice.GetVideoInputIndex()
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeInt32, data)
		case MetadataCameraStatus:
			queryParams, edgexErr := getQueryParameters(req)
			if edgexErr != nil {
				return responses, errors.NewCommonEdgeXWrapper(edgexErr)
			}

			index := queryParams.Get(InputIndex)
			if len(index) == 0 {
				return responses, fmt.Errorf("mandatory query parameter %s not found", InputIndex)
			}
			data, err = getInputStatus(cameraDevice, index)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeUint32, data)
		case MetadataImageFormats:
			data, err = getImageFormats(cameraDevice)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
		case MetadataDataFormat:
			data, err = getDataFormat(cameraDevice)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
		case MetadataCroppingAbility:
			data, err = getCropInfo(cameraDevice)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
		case MetadataStreamingParameters:
			data, err = getStreamingParameters(cameraDevice)
			if err != nil {
				return responses, errorWrapper.CommandError(command, err)
			}
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, data)
		case VideoStreamUri:
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, req.Type, device.rtspUri)
		case VideoStreamingStatus:
			cv, err = sdkModels.NewCommandValue(req.DeviceResourceName, common.ValueTypeObject, device.streamingStatus)
		default:
			return responses, errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported command %s", command), nil)
		}
		if err != nil {
			return responses, errors.NewCommonEdgeX(errors.KindServerError, "failed to create CommandValue", err)
		}
		responses[i] = cv
	}

	return responses, nil
}

func (d *Driver) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties,
	reqs []sdkModels.CommandRequest, params []*sdkModels.CommandValue) error {
	device, edgexErr := d.getDevice(deviceName)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}

	for i, req := range reqs {
		command, ok := req.Attributes[Command]
		if !ok {
			return errors.NewCommonEdgeX(errors.KindContractInvalid,
				fmt.Sprintf("command for USB camera resource %s is not specified, please check device profile",
					req.DeviceResourceName), nil)
		}
		switch command {
		case VideoStartStreaming:
			options, edgexErr := params[i].ObjectValue()
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
			device.StopStreaming()
		default:
			return errors.NewCommonEdgeX(errors.KindContractInvalid, fmt.Sprintf("unsupported command %s", command), nil)
		}
		go d.publishStreamingStatus(device)
	}

	return nil
}

func (d *Driver) Stop(force bool) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	d.wg.Add(len(d.activeDevices))

	// The wait group is used here as well as in the startStreaming functions.
	// The call to Wait() waits for StopStreaming to return and startStreaming to end.
	defer d.wg.Wait()

	for _, device := range d.activeDevices {
		go func(device *Device) {
			device.StopStreaming()
			d.wg.Done()
		}(device)
	}
	return nil
}

// AddDevice is a callback function that is invoked
// when a new Device associated with this Device Service is added
func (d *Driver) AddDevice(deviceName string, protocols map[string]models.ProtocolProperties,
	adminState models.AdminState) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	_, sn, err := getUSBDeviceIdInfo(protocols[UsbProtocol][Path])
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("could not find the serial number of the device %s", deviceName), err)
	}
	for _, ad := range d.activeDevices {
		if ad.serialNumber == sn {
			return errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("the serial number %s conflicts with existing device %s", sn, ad.name), nil)
		}
	}
	activeDevice, edgexErr := d.newDevice(deviceName, protocols)
	if edgexErr != nil {
		return errors.NewCommonEdgeXWrapper(edgexErr)
	}
	d.activeDevices[deviceName] = activeDevice
	d.lc.Debugf("a new Device is added: %s", deviceName)
	if activeDevice.autoStreaming {
		edgexErr = d.startStreaming(activeDevice)
		if edgexErr != nil {
			return errors.NewCommonEdgeXWrapper(edgexErr)
		}
	}
	return nil
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
		device.StopStreaming()
		delete(d.activeDevices, deviceName)
		d.lc.Debugf("Device %s is removed", deviceName)
	}
	return nil
}

// RefreshExistingDevicePaths checks whether the existing devices match the connected devices.
// If there is a mismatch between them, scan all paths to find the matching device and update the existing device with the correct path.
func (d *Driver) RefreshExistingDevicePaths() {
	for _, cd := range d.ds.Devices() {
		fdPath := cd.Protocols[UsbProtocol][Path]
		cn, sn, err := getUSBDeviceIdInfo(fdPath)
		if err != nil {
			d.lc.Errorf("failed to get the serial number of device %s, error: %s", cd.Name, err.Error())
		}
		// If the card name or serial number is different, it means that the path of the device has changed.
		if cn != cd.Protocols[UsbProtocol][CardName] || sn != cd.Protocols[UsbProtocol][SerialNumber] {
			go d.updateDevicePath(cd)
		}
	}
}

// Discover triggers protocol specific device discovery, which is an asynchronous operation.
// Devices found as part of this discovery operation are written to the channel devices.
func (d *Driver) Discover() {
	var devices []sdkModels.DiscoveredDevice
	// Convert the slice of cached devices to map in order to improve the performance in the subsequent for loop.
	currentDevices := d.cachedDeviceMap()
	// The file descriptor of video capture device can be /dev/video0 ~ 63
	// https://github.com/torvalds/linux/blob/master/Documentation/admin-guide/devices.txt#L1402-L1406
	for i := 0; i < 64; i++ {
		fdPath := BasePath + strconv.Itoa(i)
		if ok := d.isVideoCaptureDevice(fdPath); ok {
			cn, sn, err := getUSBDeviceIdInfo(fdPath)
			if err != nil {
				d.lc.Errorf("failed to get device serial number, error: %s", err.Error())
				continue
			}
			if _, ok := currentDevices[cn+sn]; ok {
				continue
			}
			discovered := sdkModels.DiscoveredDevice{
				Name: buildDeviceName(cn, sn),
				Protocols: map[string]models.ProtocolProperties{
					UsbProtocol: {
						Path:         fdPath,
						SerialNumber: sn,
						CardName:     cn,
					},
				},
				Description: fmt.Sprintf("USB camera %s", cn),
				Labels:      []string{"auto-discovery", cn},
			}
			devices = append(devices, discovered)
			d.lc.Infof("discovered device: %s", discovered.Name)
		}
	}
	d.deviceCh <- devices
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
	return value, nil
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

	fdPath, edgexErr := d.getProtocolProperty(protocols, UsbProtocol, Path)
	if edgexErr != nil {
		return nil, errors.NewCommonEdgeXWrapper(edgexErr)
	}

	rtspUri := &url.URL{
		Scheme: RtspUriScheme,
		Host:   fmt.Sprintf("%s:%s", d.rtspHostName, d.rtspTcpPort),
	}
	rtspUri.Path = path.Join(Stream, name)

	// Create new instance of transcoder
	trans := new(transcoder.Transcoder)
	// Initialize transcoder passing the input path and output path
	err = trans.Initialize(fdPath, rtspUri.String())
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
		command, ok := r.Attributes[Command]
		if ok && command == VideoStreamingStatus {
			streamingStatusResourceName = r.Name
			break
		}
	}
	if len(streamingStatusResourceName) == 0 {
		d.lc.Warnf("there is no device resource representing StreamingStatus of the device %s, so the StreamingStatus won't be published automatically", name)
	}

	cameraDevice, err := usbdevice.Open(fdPath)
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
	psn := protocols[UsbProtocol][SerialNumber]
	pcn := protocols[UsbProtocol][CardName]
	// pre-defined devices may not include serial number information
	if len(psn) == 0 {
		device.Protocols[UsbProtocol][SerialNumber] = sn
		c, err := cameraDevice.GetCapability()
		if err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("failed to get device capability info,path=%s", fdPath), err)
		}
		device.Protocols[UsbProtocol][CardName] = c.Card
		if err := d.ds.UpdateDevice(device); err != nil {
			return nil, errors.NewCommonEdgeX(errors.KindServerError,
				fmt.Sprintf("failed to update the device %s to add serial number information", device.Name), err)
		}
	} else if pcn != cn {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("wrong device card name, expected %s=%s, actual %s=%s", CardName, pcn, CardName, cn), nil)
	} else if psn != sn {
		return nil, errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("wrong device serial number, expected %s=%s, actual %s=%s", SerialNumber, psn, SerialNumber, sn), nil)
	}

	return &Device{
		name:                        name,
		path:                        fdPath,
		serialNumber:                sn,
		rtspUri:                     rtspUri.String(),
		transcoder:                  trans,
		autoStreaming:               autoStreaming,
		streamingStatusResourceName: streamingStatusResourceName,
	}, nil
}

// getDevice gets an active device by name, which is managed by device service.
func (d *Driver) getDevice(name string) (*Device, errors.EdgeX) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	device, ok := d.activeDevices[name]
	if ok {
		return device, nil
	} else {
		return nil, errors.NewCommonEdgeX(errors.KindEntityDoesNotExist, fmt.Sprintf("device %s not found", name), nil)
	}
}

func (d *Driver) startStreaming(device *Device) errors.EdgeX {
	ctx, cancel := context.WithCancel(context.TODO())
	errChan, err := device.StartStreaming(ctx, cancel)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf(
			"failed to start video streaming for device %s", device.name), err)
	}
	d.wg.Add(1)
	go func() {
		select {
		case err := <-errChan:
			device.StopStreaming()
			d.lc.Errorf("the video streaming process for device %s has stopped, error: %s", device.name, err)
			d.wg.Done()
			return
		case <-device.ctx.Done():
			if err := device.transcoder.Stop(); err != nil {
				d.lc.Errorf("failed to stop video streaming for device %s, error: %s", device.name, err)
			}
			d.lc.Debugf("the video streaming process for device %s has stopped", device.name)
			d.wg.Done()
			return
		}
	}()
	d.lc.Infof("start video streaming for device %s", device.name)
	return nil
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
		cn := cd.Protocols[UsbProtocol][CardName]
		sn := cd.Protocols[UsbProtocol][SerialNumber]
		if len(cn) > 0 && len(sn) > 0 {
			cdm[cn+sn] = cd
		}
	}
	return cdm
}

func (d *Driver) isVideoCaptureDevice(path string) bool {
	cameraDevice, err := usbdevice.Open(path)
	if err != nil {
		d.lc.Debugf("there is no USB camera at specified path %s, error: %s", path, err.Error())
		return false
	}
	defer cameraDevice.Close()
	c, err := cameraDevice.GetCapability()
	if err != nil {
		d.lc.Errorf("failed to get device capability, path=%s, error:%s", path, err.Error())
		return false
	}
	return isVideoCaptureSupported(c) && isStreamingSupported(c)
}

func (d *Driver) updateDevicePath(device models.Device) {
	// Scan all paths to find the matching device.
	// The file descriptor of video capture device can be /dev/video0 ~ 63
	// https://github.com/torvalds/linux/blob/master/Documentation/admin-guide/devices.txt#L1402-L1406
	for i := 0; i < 64; i++ {
		fdPath := BasePath + strconv.Itoa(i)
		if ok := d.isVideoCaptureDevice(fdPath); ok {
			cn, sn, err := getUSBDeviceIdInfo(fdPath)
			if err != nil {
				d.lc.Errorf("failed to get device serial number, path=%s, error: %s", fdPath, err.Error())
				continue
			}
			if cn == device.Protocols[UsbProtocol][CardName] && sn == device.Protocols[UsbProtocol][SerialNumber] {
				device.Protocols[UsbProtocol][Path] = fdPath
				if err := d.ds.UpdateDevice(device); err != nil {
					d.lc.Errorf("failed to update path for the device %s", device.Name)
				}
				break
			}
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
