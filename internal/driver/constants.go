// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

const (
	Command                         = "command"
	UsbProtocol                     = "USB"
	Paths                           = "Paths"
	SerialNumber                    = "SerialNumber"
	CardName                        = "CardName"
	AutoStreaming                   = "AutoStreaming"
	InputIndex                      = "InputIndex"
	UrlRawQuery                     = "urlRawQuery"
	RtspServerHostName              = "RtspServerHostName"
	DefaultRtspServerHostName       = "localhost"
	RtspTcpPort                     = "RtspTcpPort"
	DefaultRtspTcpPort              = "8554"
	RtspAuthenticationServer        = "RtspAuthenticationServer"
	DefaultRtspAuthenticationServer = "localhost:8000"
	RtspUriScheme                   = "rtsp"
	Stream                          = "stream"
	PrefixInput                     = "Input"
	PrefixOutput                    = "Output"

	// API route specific to Device Service
	ApiRefreshDevicePaths = "/refreshdevicepaths"

	// Metadata descriptions
	DescNotSpecified = "not specified"
	DescTimePerFrame = "time per frame"
	DescHighQuality  = "high quality"

	// Command names
	MetadataDeviceCapability    = "METADATA_DEVICE_CAPABILITY"
	MetadataCurrentVideoInput   = "METADATA_CURRENT_VIDEO_INPUT"
	MetadataCameraStatus        = "METADATA_CAMERA_STATUS"
	MetadataDataFormat          = "METADATA_DATA_FORMAT"
	MetadataCroppingAbility     = "METADATA_CROPPING_ABILITY"
	MetadataStreamingParameters = "METADATA_STREAMING_PARAMETERS"
	MetadataImageFormats        = "METADATA_IMAGE_FORMATS"
	VideoStartStreaming         = "VIDEO_START_STREAMING"
	VideoStopStreaming          = "VIDEO_STOP_STREAMING"
	VideoStreamUri              = "VIDEO_STREAM_URI"
	VideoStreamingStatus        = "VIDEO_STREAMING_STATUS"
	VideoSetFPS                 = "VIDEO_SET_FPS"

	// FFmpeg options
	FFmpegFrames      = "-frames:d"
	FFmpegFps         = "-r"
	FFmpegSize        = "-s"
	FFmpegAspect      = "-aspect"
	FFmpegQScale      = "-qscale"
	FFmpegVCodec      = "-vcodec"
	FFmpegInputFormat = "-input_format"

	// FFmpeg option values
	FFmpegPixelFmtRGB24 = "rgb24"
	FFmpegPixelFmtGray  = "gray"
	FFmpegPixelFmtYUYV  = "yuyv422"
	FFmpegPixelFmtMJPEG = "mjpeg"

	// Input option names
	InputPixelFormat = "InputPixelFormat"

	// udev device properties
	UdevSerialShort = "ID_SERIAL_SHORT"
	UdevSerial      = "ID_SERIAL"
	UdevV4lProduct  = "ID_V4L_PRODUCT"
)
