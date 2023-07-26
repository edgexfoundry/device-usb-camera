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
