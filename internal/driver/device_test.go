// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBuildDeviceName(t *testing.T) {
	tests := []struct {
		name     string
		serial   string
		expected string
	}{
		{
			name:     "UVC Camera (046d:0825)",
			serial:   "61C0AE50",
			expected: "UVC_Camera_046d_0825-61C0AE50",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, buildDeviceName(test.name, test.serial))
		})
	}
}
