// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vladimirvivien/go4vl/v4l2"
)

func TestParseOptionValuePixelFormat(t *testing.T) {
	tests := []struct {
		name          string
		value         interface{}
		expectedValue string
		expectErr     bool
	}{
		{"rgb24 (go4vl)", v4l2.PixelFormats[v4l2.PixelFmtRGB24], FFmpegPixelFmtRGB24, false},
		{"rgb24 (FFmpeg)", FFmpegPixelFmtRGB24, FFmpegPixelFmtRGB24, false},
		{"gray (go4vl)", v4l2.PixelFormats[v4l2.PixelFmtGrey], FFmpegPixelFmtGray, false},
		{"gray (FFmpeg)", FFmpegPixelFmtGray, FFmpegPixelFmtGray, false},
		{"yuyv (go4vl)", v4l2.PixelFormats[v4l2.PixelFmtYUYV], FFmpegPixelFmtYUYV, false},
		{"yuyv (FFmpeg)", FFmpegPixelFmtYUYV, FFmpegPixelFmtYUYV, false},
		{"mjpeg (go4vl)", v4l2.PixelFormats[v4l2.PixelFmtMJPEG], FFmpegPixelFmtMJPEG, false},
		{"mjpeg (FFmpeg)", FFmpegPixelFmtMJPEG, FFmpegPixelFmtMJPEG, false},
		{"unsupported value", "rgb8", "", true},
		{"wrong value type", 123, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := parseOptionValue(InputPixelFormat, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.Equal(t, value, tt.expectedValue)
			}
		})
	}
}
