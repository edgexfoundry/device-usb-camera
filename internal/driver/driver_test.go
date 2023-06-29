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
				paths: []{
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
			CardName:     "testDevice" + strconv.Itoa(a),
			SerialNumber: strconv.Itoa(a) + strconv.Itoa(b) + strconv.Itoa(c),
			Paths: []interface{}{
				"/dev/video" + strconv.Itoa(a),
				"/dev/video" + strconv.Itoa(b),
				"/dev/video" + strconv.Itoa(c),
			},
		},
	}}
}

func TestDriver_cachedDeviceMap(t *testing.T) {
	tests := []struct {
		name     string
		devices  []models.Device
		expected map[string]models.Device
	}{
		{
			name: "happy path single device",
			devices: []models.Device{
				createTestDevice(0, 2, 4),
			},
			expected: map[string]models.Device{
				"testDevice0024": createTestDevice(0, 2, 4),
			},
		},
		{
			name: "happy path multiple devices",
			devices: []models.Device{
				createTestDevice(0, 2, 4),
				createTestDevice(6, 8, 10),
				createTestDevice(12, 14, 16),
				createTestDevice(18, 20, 22),
			},
			expected: map[string]models.Device{
				"testDevice0024":     createTestDevice(0, 2, 4),
				"testDevice66810":    createTestDevice(6, 8, 10),
				"testDevice12121416": createTestDevice(12, 14, 16),
				"testDevice18182022": createTestDevice(18, 20, 22),
			},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			driver, mockService := createDriverWithMockService()
			mockService.On("Devices").
				Return(test.devices)
			cacheMap := driver.cachedDeviceMap()
			assert.Equal(t, test.expected, cacheMap)
		})
	}
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
		expected      []string
		errorExpected bool
	}{
		{
			name:   "happy path",
			device: createTestDevice(0, 2, 4),
			expected: []string{
				"/dev/video0",
				"/dev/video2",
				"/dev/video4",
			},
		},
		{
			name: "empty paths",
			device: models.Device{
				Name: "testDeviceRealsense",
				Protocols: map[string]models.ProtocolProperties{
					UsbProtocol: map[string]any{
						Paths: []interface{}{},
					},
				},
			},
			expected: []string{},
		},
		{
			name: "no paths field",
			device: models.Device{
				Name: "testDeviceRealsense",
				Protocols: map[string]models.ProtocolProperties{
					UsbProtocol: map[string]any{},
				},
			},
			errorExpected: true,
		},
		{
			name: "no Protocols field",
			device: models.Device{
				Name:      "testDeviceRealsense",
				Protocols: map[string]models.ProtocolProperties{},
			},
			errorExpected: true,
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
