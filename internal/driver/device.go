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
	serialNumber                string
	rtspUri                     string
	transcoder                  *transcoder.Transcoder
	autoStreaming               bool
	mutex                       sync.Mutex
	streamingStatus             streamingStatus
	streamingStatusResourceName string
}

type streamingStatus struct {
	IsStreaming        bool
	Error              string
	OutputFrames       string
	InputFps           string
	OutputFps          string
	InputImageSize     string
	OutputImageSize    string
	OutputAspect       string
	OutputVideoQuality string
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

// SetFps updates the fps on the device side of the service. Note that this won't update the rtsp output stream fps
func (dev *Device) SetFps(device *usbdevice.Device, fpsNumerator uint32, fpsDenominator uint32) (string, error) {
	fps := fmt.Sprintf("%f", float32(fpsDenominator)/float32(fpsNumerator))
	dataFormat, err := getDataFormat(device)
	if err != nil {
		return "", nil
	}
	intervals := dataFormat.(DataFormat).FpsIntervals
	foundFlag := false
	for _, interval := range intervals {
		if fpsNumerator == interval.Numerator && fpsDenominator == interval.Denominator {
			foundFlag = true
			break
		}
	}
	if !foundFlag {
		return "", errors.NewCommonEdgeX(errors.KindCommunicationError, fmt.Sprintf("FPS value %s not supported for current image format.", fps), nil)
	}

	// Update device fps for stream parameters
	origStreamParam, err := device.GetStreamParam()
	if err != nil {
		return "", err
	}
	origStreamParam.Capture.TimePerFrame.Denominator = fpsDenominator
	origStreamParam.Capture.TimePerFrame.Numerator = fpsNumerator
	err = device.SetStreamParam(origStreamParam)
	if err != nil {
		return "", err
	}

	return fps, nil
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
