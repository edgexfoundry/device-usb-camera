// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import "github.com/vladimirvivien/go4vl/v4l2"

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

type PixelFormat struct {
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

var PixelFormatPixelFormats = map[string]uint32{
	"RGB":   v4l2.PixelFmtRGB24,
	"GREY":  v4l2.PixelFmtGrey,
	"YUYV":  v4l2.PixelFmtYUYV,
	"MJPG":  v4l2.PixelFmtMJPEG,
	"JPEG":  v4l2.PixelFmtJPEG,
	"MPEG":  v4l2.PixelFmtMPEG,
	"H264":  v4l2.PixelFmtH264,
	"MPEG4": v4l2.PixelFmtMPEG4,
	"UYVY":  v4l2.PixelFmtUYVY,
	// pixel formats not directly included as part of v4l2 pixel format definitions
	"BYR2": 844257602,
	"Z16":  540422490,
	"Y8I":  541669465,
	"Y12I": 1228026201,
}
