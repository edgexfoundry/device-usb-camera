// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"strconv"
	"testing"

	sdkMocks "github.com/edgexfoundry/device-sdk-go/v3/pkg/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
)

func createDriverWithMockService() (*Driver, *sdkMocks.DeviceServiceSDK) {
	mockService := &sdkMocks.DeviceServiceSDK{}
	driver := NewDriver()
	driver.lc = logger.MockLogger{}
	driver.ds = mockService
	mockService.On("LoggingClient").Return(driver.lc).Maybe()
	return driver, mockService
}

func NewDriver() *Driver {
	return &Driver{
		activeDevices: map[string]*Device{
			"testDeviceRealsense": &Device{
				paths: []interface{}{
					"/dev/video0",
					"/dev/video2",
					"/dev/video4",
				},
			},
		},
	}
}

func createTestDevice(a, b, c int) models.Device {
	return models.Device{Name: "testDeviceRealsense", Protocols: map[string]models.ProtocolProperties{
		UsbProtocol: map[string]any{
			Paths: []interface{}{
				"/dev/video" + strconv.Itoa(a),
				"/dev/video" + strconv.Itoa(b),
				"/dev/video" + strconv.Itoa(c),
			},
		},
	}}
}

// func TestDriver_updateDevicePath(t *testing.T) {
// 	driver, mockService := createDriverWithMockService()
// 	testDevice := createTestDevice()

// 	mockService.On("isVideoCaptureDevice", "/dev/video0").
// 		Return(true)

// 	mockService.On("UpdateDevice", testDevice).
// 		Return(nil)

// 	driver.updateDevicePath(testDevice)
// }

func TestDriver_getPaths(t *testing.T) {
	tests := []struct {
		name          string
		device        models.Device
		expected      []interface{}
		errorExpected bool
	}{
		{
			name:   "happyPath",
			device: createTestDevice(0, 2, 4),
			expected: []interface{}{
				"/dev/video0",
				"/dev/video2",
				"/dev/video4",
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			paths, err := driver.getPaths(test.device.Protocols)
			if test.errorExpected {
				require.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, test.expected, paths)
		})
	}
}
