// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

const (
	Command                   = "command"
	UsbProtocol               = "USB"
	Path                      = "Path"
	SerialNumber              = "SerialNumber"
	CardName                  = "CardName"
	BasePath                  = "/dev/video"
	AutoStreaming             = "AutoStreaming"
	InputIndex                = "InputIndex"
	UrlRawQuery               = "urlRawQuery"
	RtspServerHostName        = "RtspServerHostName"
	DefaultRtspServerHostName = "localhost"
	RtspTcpPort               = "RtspTcpPort"
	DefaultRtspTcpPort        = "8554"
	RtspUriScheme             = "rtsp"
	Stream                    = "stream"
	PrefixInput               = "Input"
	PrefixOutput              = "Output"

	// API route specific to Device Service
	ApiRefreshDevicePaths = "/refreshdevicepaths"

	// Metadata descriptions
	DescNotSpecified = "not specified"
	DescTimePerFrame = "time per frame"
	DescHighQuality  = "high quality"

	// V4L2 commands
	VIDIOC_QUERYCAP  = "VIDIOC_QUERYCAP"
	VIDIOC_G_INPUT   = "VIDIOC_G_INPUT"
	VIDIOC_ENUMINPUT = "VIDIOC_ENUMINPUT"
	VIDIOC_G_FMT     = "VIDIOC_G_FMT"
	VIDIOC_CROPCAP   = "VIDIOC_CROPCAP"
	VIDIOC_G_PARM    = "VIDIOC_G_PARM"
	VIDIOC_ENUM_FMT  = "VIDIOC_ENUM_FMT"

	// EdgeX commands
	EDGEX_START_STREAMING  = "EDGEX_START_STREAMING"
	EDGEX_STOP_STREAMING   = "EDGEX_STOP_STREAMING"
	EDGEX_STREAM_URI       = "EDGEX_STREAM_URI"
	EDGEX_STREAMING_STATUS = "EDGEX_STREAMING_STATUS"

	// FFmpeg options
	FFmpegFrames      = "-frames:d"
	FFmpegFps         = "-r"
	FFmpegSize        = "-s"
	FFmpegAspect      = "-aspect"
	FFmpegQScale      = "-qscale"
	FFmpegVCodec      = "-vcodec"
	FFmpegInputFormat = "-input_format"

	// udev device properties
	UdevSerialShort = "ID_SERIAL_SHORT"
	UdevV4lProduct  = "ID_V4L_PRODUCT"
)
