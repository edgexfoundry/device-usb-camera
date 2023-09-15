// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

const (
	GetFunction                     = "getFunction"
	SetFunction                     = "setFunction"
	UsbProtocol                     = "USB"
	Paths                           = "Paths"
	Path                            = "Path"
	SerialNumber                    = "SerialNumber"
	CardName                        = "CardName"
	AutoStreaming                   = "AutoStreaming"
	InputIndex                      = "InputIndex"
	UrlRawQuery                     = "urlRawQuery"
	RtspServerMode                  = "RtspServerMode"
	RtspServerExe                   = "RtspServerExecutable"
	RtspServerExeDefault            = "./rtsp-simple-server"
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
	FrameRateValueDenominator       = "FrameRateValueDenominator"
	FrameRateValueNumerator         = "FrameRateValueNumerator"
	PathIndex                       = "PathIndex"
	Width                           = "Width"
	Height                          = "Height"
	PixelFormat                     = "PixelFormat"
	StreamFormat                    = "StreamFormat"
	RGB                             = "RGB"
	Greyscale                       = "Greyscale"
	Depth                           = "Depth"

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
	MetadataFrameRateFormats    = "METADATA_FRAMERATE_FORMATS"
	VideoStartStreaming         = "VIDEO_START_STREAMING"
	VideoStopStreaming          = "VIDEO_STOP_STREAMING"
	VideoStreamUri              = "VIDEO_STREAM_URI"
	VideoStreamingStatus        = "VIDEO_STREAMING_STATUS"
	VideoGetFrameRate           = "VIDEO_GET_FRAMERATE"
	VideoSetFrameRate           = "VIDEO_SET_FRAMERATE"
	VideoGetPixelFormat         = "VIDEO_GET_PIXELFORMAT"
	VideoSetPixelFormat         = "VIDEO_SET_PIXELFORMAT"

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

	// Pixel Formats not supported by go4vl pre-defined pixel format definitions
	PixFmtBYR2     = 844257602
	PixFmtDepthZ16 = 540422490
	PixFmtY8I      = 541669465
	PixFmtY12I     = 1228026201

	RedactedStr = "//<redacted>@"

	RTSPServerModeInternal RTSPServerMode = "internal"
	RTSPServerModeExternal RTSPServerMode = "external"
	RTSPServerModeNone     RTSPServerMode = "none"

	// RtspAuthSecretName defines the secretName used for storing RTSP credentials in the secret store.
	RtspAuthSecretName string = "rtspauth"
)
