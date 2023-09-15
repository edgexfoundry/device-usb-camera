// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import "github.com/vladimirvivien/go4vl/v4l2"

type RTSPServerMode string

type RTSPAuthRequest struct {
	IP       string `json:"ip"`
	User     string `json:"user"`
	Password string `json:"password"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	ID       string `json:"id"`
	Action   string `json:"action"`
	Query    string `json:"query"`
}

type FrameInfo struct {
	Index       uint32
	FrameType   uint32
	PixelFormat uint32
	Height      uint32
	Width       uint32
	Rates       []v4l2.Fract
}

type FrameRateFormat struct {
	Description string
	FrameRates  []FrameInfo
}

type VideoPixelFormat struct {
	Width        uint32 `json:"Width"`
	Height       uint32 `json:"Height"`
	PixelFormat  string `json:"PixelFormat"`
	Field        string `json:"Field"`
	BytesPerLine uint32 `json:"BytesPerLine"`
	SizeImage    uint32 `json:"SizeImage"`
	Colorspace   string `json:"Colorspace"`
	Priv         uint32 `json:"Priv"`
	Flags        uint32 `json:"Flags"`
	YcbcrEnc     string `json:"YcbcrEnc"`
	HSVEnc       string `json:"HSVEnc"`
	Quantization string `json:"Quantization"`
	XferFunc     string `json:"XferFunc"`
}

type StreamingStatus struct {
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

var PixelFormatV4l2Mappings = map[string]uint32{
	"RGB":   v4l2.PixelFmtRGB24,
	"GREY":  v4l2.PixelFmtGrey,
	"YUYV":  v4l2.PixelFmtYUYV,
	"MJPG":  v4l2.PixelFmtMJPEG,
	"JPEG":  v4l2.PixelFmtJPEG,
	"MPEG":  v4l2.PixelFmtMPEG,
	"H264":  v4l2.PixelFmtH264,
	"MPEG4": v4l2.PixelFmtMPEG4,
	"UYVY":  v4l2.PixelFmtUYVY,
	// pixel formats not supported by go4vl pixel format definitions
	"BYR2": PixFmtBYR2,
	"Z16":  PixFmtDepthZ16,
	"Y8I":  PixFmtY8I,
	"Y12I": PixFmtY12I,
}

var StreamFormatTypeMap = map[uint32]string{
	v4l2.PixelFmtRGB24: RGB,
	v4l2.PixelFmtGrey:  Greyscale,
	v4l2.PixelFmtYUYV:  RGB,
	v4l2.PixelFmtMJPEG: RGB,
	v4l2.PixelFmtJPEG:  RGB,
	v4l2.PixelFmtMPEG:  RGB,
	v4l2.PixelFmtH264:  RGB,
	v4l2.PixelFmtMPEG4: RGB,
	v4l2.PixelFmtUYVY:  Greyscale,
	PixFmtBYR2:         RGB,
	PixFmtY8I:          Greyscale,
	PixFmtY12I:         Greyscale,
	PixFmtDepthZ16:     Depth,
}
