// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"

	usbdevice "github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"

	"github.com/xfrr/goffmpeg/transcoder"
)

const (
	// Per https://tools.ietf.org/html/rfc3986#section-2.3, unreserved characters = ALPHA / DIGIT / "-" / "." / "_" / "~"
	// Also due to names used in topics for Redis Pub/Sub, "." are not allowed
	// see: https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/common/validator.go
	//
	// Note: this is an inverted match of the unreserved characters from above, as we want to remove the reserved ones
	rFC3986ReservedCharsRegexString = "[^a-zA-Z0-9-_~]+"
)

var (
	rFC3986ReservedCharsRegex = regexp.MustCompile(rFC3986ReservedCharsRegexString)
)

type Device struct {
	lc                          logger.LoggingClient
	name                        string
	paths                       []string
	path                        string
	serialNumber                string
	rtspUri                     string
	transcoder                  *transcoder.Transcoder
	autoStreaming               bool
	mutex                       sync.Mutex
	streamingStatus             streamingStatus
	streamingStatusResourceName string
}

type streamingStatus struct {
	TranscoderInputPath string
	IsStreaming         bool
	Error               string
	OutputFrames        string
	InputFps            string
	OutputFps           string
	InputImageSize      string
	OutputImageSize     string
	OutputAspect        string
	OutputVideoQuality  string
}

func (dev *Device) StartStreaming() (<-chan string, <-chan error, error) {
	dev.mutex.Lock()
	isStreaming := dev.streamingStatus.IsStreaming
	dev.mutex.Unlock()
	if isStreaming {
		return nil, nil, fmt.Errorf("video streaming is already in progress")
	}

	dev.lc.Infof("Attempting to start streaming device %s", dev.name)
	progressChan, errChan, err := dev.runTranscoderWithOutput()
	if err != nil {
		wrappedErr := errors.NewCommonEdgeX(errors.KindServerError, "failed running ffmpeg transcoder for device "+dev.name, err)
		return nil, nil, wrappedErr
	}
	return progressChan, errChan, nil
}

func (dev *Device) StopStreaming() {
	dev.mutex.Lock()
	defer dev.mutex.Unlock()
	if !dev.streamingStatus.IsStreaming {
		return
	}

	dev.lc.Debugf("Stopping transcoder for device %s", dev.name)
	if err := dev.transcoder.Stop(); err != nil {
		dev.lc.Errorf("Failed to stop video streaming transcoder for device %s, error: %s", dev.name, err)
		return
	}
}

func (dev *Device) updateTranscoderInputPath(fdPath string) error {
	trans := dev.transcoder
	err := trans.SetInputPath(fdPath)
	if err != nil {
		return errors.NewCommonEdgeX(errors.KindServerError,
			fmt.Sprintf("failed to set new path for transcoder for device %s", dev.name), err)
	}
	dev.streamingStatus.TranscoderInputPath = fdPath
	dev.lc.Debugf("Transcoder path succesfully set to %s", fdPath)
	return nil
}

func (dev *Device) SetPixelFormat(usbDevice *usbdevice.Device, params interface{}) error {
	// Get the current video pixel format to populate the fields missing from the input
	v4l2PixelFormat, err := usbDevice.GetPixFormat()
	if err != nil {
		dev.lc.Errorf("error getting current pixel format for the device %s, error: %s", dev.name, err)
		return err
	}

	widthValue, ok := params.(map[string]interface{})[Width]
	if ok {
		width, err := strconv.ParseUint(widthValue.(string), 0, 32)
		if err != nil {
			return fmt.Errorf("invalid input: error parsing width for the device %s, error: %s", dev.name, err)
		}
		v4l2PixelFormat.Width = uint32(width)
	}

	heightValue, ok := params.(map[string]interface{})[Height]
	if ok {
		height, err := strconv.ParseUint(heightValue.(string), 0, 32)
		if err != nil {
			return fmt.Errorf("invalid input: error parsing height for the device %s, error: %s", dev.name, err)
		}
		v4l2PixelFormat.Height = uint32(height)
	}

	pixFormatValue, ok := params.(map[string]interface{})[PixelFormat]
	if ok {
		pixelFormat, ok := PixelFormatV4l2Mappings[pixFormatValue.(string)]
		if !ok {
			return fmt.Errorf("invalid input: error parsing pixelFormat for the device %s, error: %s", dev.name, err)
		}

		// Check if the given pixelFormat is supported for the device video streaming path
		supported, err := isPixFormatSupported(pixelFormat, usbDevice)
		if err != nil {
			return err
		}
		if supported {
			v4l2PixelFormat.PixelFormat = pixelFormat
		} else {
			return fmt.Errorf("invalid input: pixelFormat for the given path not supported by the device %s", dev.name)
		}
	}

	err = usbDevice.SetPixFormat(v4l2PixelFormat)
	if err != nil {
		dev.lc.Errorf("error setting pixel format for the device %s, error: %s", dev.name, err)
		return err
	}

	return nil
}

// SetFrameRate updates the fps on the device side of the service. Note that this won't update the rtsp output stream fps
func (dev *Device) SetFrameRate(usbDevice *usbdevice.Device, frameRateNumerator uint32, frameRateDenominator uint32) (string, error) {
	fps := fmt.Sprintf("%f", float32(frameRateNumerator)/float32(frameRateDenominator))
	dataFormat, err := getDataFormat(usbDevice)
	if err != nil {
		return "", err
	}
	found := false
	for _, format := range dataFormat.(map[string]DataFormat) {
		for _, frameRate := range format.FrameRates {
			if frameRateNumerator == frameRate.Numerator && frameRateDenominator == frameRate.Denominator {
				found = true
				break
			}
		}
	}
	if !found {
		return "", errors.NewCommonEdgeX(errors.KindCommunicationError, fmt.Sprintf("FPS value %s not supported for current image format.", fps), nil)
	}

	// Update device fps for stream parameters
	origStreamParam, err := usbDevice.GetStreamParam()
	if err != nil {
		return "", err
	}
	// this swaps user-friendly frame rate (frames per second) to
	// the internally tracked frame interval (seconds per frame)
	origStreamParam.Capture.TimePerFrame.Denominator = frameRateNumerator
	origStreamParam.Capture.TimePerFrame.Numerator = frameRateDenominator
	err = usbDevice.SetStreamParam(origStreamParam)
	if err != nil {
		return "", err
	}

	return fps, nil
}

func (dev *Device) GetFrameRate(usbDevice *usbdevice.Device) (v4l2.Fract, error) {
	streamParam, err := usbDevice.GetStreamParam()
	if err != nil {
		return v4l2.Fract{}, err
	}
	timePerFrame := streamParam.Capture.TimePerFrame
	var fps v4l2.Fract
	fps.Denominator = timePerFrame.Numerator
	fps.Numerator = timePerFrame.Denominator
	return fps, nil
}

func (dev *Device) GetPixelFormat(usbDevice *usbdevice.Device) (interface{}, error) {
	pixFmt, err := usbDevice.GetPixFormat()
	if err != nil {
		return nil, err
	}

	result := VideoPixelFormat{
		Width:        pixFmt.Width,
		Height:       pixFmt.Height,
		PixelFormat:  v4l2.PixelFormats[pixFmt.PixelFormat],
		Field:        v4l2.Fields[pixFmt.Field],
		BytesPerLine: pixFmt.BytesPerLine,
		SizeImage:    pixFmt.SizeImage,
		Colorspace:   v4l2.Colorspaces[pixFmt.Colorspace],
		Priv:         pixFmt.Priv,
		Flags:        pixFmt.Flags,
		YcbcrEnc:     v4l2.YCbCrEncodings[pixFmt.YcbcrEnc],
		HSVEnc:       v4l2.YCbCrEncodings[pixFmt.HSVEnc],
		Quantization: v4l2.Quantizations[pixFmt.Quantization],
		XferFunc:     v4l2.XferFunctions[pixFmt.XferFunc],
	}

	// Since the go4vl library has limited pre-defined Pixel Format descriptions
	// get the missing pixel format description using GetFormatDescriptions().
	if result.PixelFormat == "" {
		desc, err := getPixFormatDesc(usbDevice, pixFmt.PixelFormat)
		if err != nil {
			return nil, err
		}
		result.PixelFormat = desc
	}

	return result, nil
}

func (dev *Device) updateFFmpegOptions(optName, optVal string) {
	ps := reflect.ValueOf(&dev.streamingStatus)
	s := ps.Elem()
	f := s.FieldByName(optName)
	if f.IsValid() {
		f.SetString(optVal)
	}
}

func isVideoCaptureSupported(caps v4l2.Capability) bool {
	return (caps.DeviceCapabilities & v4l2.CapVideoCapture) != 0
}

func isStreamingSupported(caps v4l2.Capability) bool {
	return (caps.DeviceCapabilities & v4l2.CapStreaming) != 0
}

func buildDeviceName(cardName, serialNumber string) string {
	return fmt.Sprintf("%s-%s",
		// replace all the reserved chars with an underscore, and trim any leftovers
		strings.Trim(rFC3986ReservedCharsRegex.ReplaceAllString(cardName, "_"), "_"),
		strings.Trim(rFC3986ReservedCharsRegex.ReplaceAllString(serialNumber, "_"), "_"))
}

func isPixFormatSupported(input uint32, d *usbdevice.Device) (bool, error) {
	pixFormats, err := d.GetFormatDescriptions()
	if err != nil {
		return false, err
	}

	for _, pix := range pixFormats {
		if input == pix.PixelFormat {
			return true, nil
		}
	}

	return false, nil
}
