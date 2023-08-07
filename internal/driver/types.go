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
}

var PixelFormatPixelFormats = map[string]uint32{
	"RGB":   v4l2.PixelFmtRGB24,
	"GREY":  v4l2.PixelFmtGrey,
	"YUYV":  v4l2.PixelFmtYUYV,
	"MJPEG": v4l2.PixelFmtMJPEG,
	"JPEG":  v4l2.PixelFmtJPEG,
	"MPEG":  v4l2.PixelFmtMPEG,
	"H264":  v4l2.PixelFmtH264,
	"MPEG4": v4l2.PixelFmtMPEG4,
}

var PixelFormatFields = map[string]uint32{
	"ANY":       v4l2.FieldAny,
	"NONE":      v4l2.FieldNone,
	"TOP":       v4l2.FieldTop,
	"BOTTOM":    v4l2.FieldBottom,
	"INT":       v4l2.FieldInterlaced,
	"SEQTOPBOT": v4l2.FieldInterlacedBottomTop,
	"SEQBOTTOP": v4l2.FieldInterlacedTopBottom,
	"ALT":       v4l2.FieldAlternate,
	"INTTOPBOT": v4l2.FieldInterlacedTopBottom,
	"INTBOTTOP": v4l2.FieldSequentialBottomTop,
}

var PixelFormatColorspaces = map[string]uint32{
	"DEFAULT": v4l2.ColorspaceDefault,
	"REC709":  v4l2.ColorspaceREC709,
	"470SBG":  v4l2.Colorspace470SystemBG,
	"JPEG":    v4l2.ColorspaceJPEG,
	"SRGB":    v4l2.ColorspaceSRGB,
	"OPRGB":   v4l2.ColorspaceOPRGB,
	"BT2020":  v4l2.ColorspaceBT2020,
	"RAW":     v4l2.ColorspaceRaw,
	"DCIP3":   v4l2.ColorspaceDCIP3,
}
