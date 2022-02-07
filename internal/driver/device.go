// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/vladimirvivien/go4vl/v4l2"
	"github.com/xfrr/goffmpeg/transcoder"
)

type Device struct {
	name                        string
	path                        string
	serialNumber                string
	rtspUri                     string
	transcoder                  *transcoder.Transcoder
	ctx                         context.Context
	cancelFunc                  context.CancelFunc
	autoStreaming               bool
	mutex                       sync.Mutex
	streamingStatus             streamingStatus
	streamingStatusResourceName string
}

type streamingStatus struct {
	IsStreaming        bool
	OutputFrames       string
	InputFps           string
	OutputFps          string
	InputImageSize     string
	OutputImageSize    string
	OutputAspect       string
	OutputVideoQuality string
}

func (dev *Device) StartStreaming(ctx context.Context, cancel context.CancelFunc) (<-chan error, error) {
	dev.mutex.Lock()
	defer dev.mutex.Unlock()
	if dev.streamingStatus.IsStreaming {
		return nil, fmt.Errorf("video streaming is already in progress")
	}
	dev.ctx = ctx
	dev.cancelFunc = cancel
	errChan := dev.transcoder.Run(false)
	dev.streamingStatus.IsStreaming = true
	return errChan, nil
}

func (dev *Device) StopStreaming() {
	dev.mutex.Lock()
	defer dev.mutex.Unlock()
	if dev.streamingStatus.IsStreaming {
		dev.cancelFunc()
		dev.streamingStatus.IsStreaming = false
	}
}

func (dev *Device) updateFFmpegOptions(optName, optVal string) {
	ps := reflect.ValueOf(&dev.streamingStatus)
	s := ps.Elem()
	f := s.FieldByName(optName)
	if f.IsValid() {
		f.SetString(optVal)
	}
}

func isVideoCaptureSupported(caps *v4l2.Capability) bool {
	return (caps.DeviceCapabilities & v4l2.CapVideoCapture) != 0
}

func isStreamingSupported(caps *v4l2.Capability) bool {
	return (caps.DeviceCapabilities & v4l2.CapStreaming) != 0
}

func buildDeviceName(cardName, serialNumber string) string {
	cardName = strings.ReplaceAll(strings.ReplaceAll(cardName, " ", "_"), ":", "_")
	serialNumber = strings.ReplaceAll(strings.ReplaceAll(serialNumber, ":", "_"), ".", "_")
	return fmt.Sprintf("%s-%s", cardName, serialNumber)
}
