// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
// Copyright (C) 2025 IOTEch Ltd
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strconv"

	usbdevice "github.com/vladimirvivien/go4vl/device"
	"github.com/vladimirvivien/go4vl/v4l2"
)

type Capability struct {
	Driver     string
	Card       string
	BusInfo    string
	Version    string
	DriverCaps []string
	DeviceCaps []string
}

type DataFormat struct {
	Width                  uint32
	Height                 uint32
	PixelFormat            uint32
	PixelFormatDescription string
	Field                  string
	BytesPerLine           uint32
	SizeImage              uint32
	Colorspace             string
	XferFunc               string
	YcbcrEnc               string
	Quantization           string
	FrameRates             []v4l2.Fract
}

type CaptureMode struct {
	Desc  string
	Value v4l2.StreamParamFlag
}

type StreamingCapability struct {
	Desc  string
	Value v4l2.StreamParamFlag
}

type StreamingParameters struct {
	Capability   StreamingCapability
	CaptureMode  CaptureMode
	TimePerFrame v4l2.Fract
	ReadBuffers  uint32
}

type ImageFormat struct {
	Index       uint32
	BufType     v4l2.BufType
	Flags       v4l2.FmtDescFlag
	Description string
	PixelFormat uint32
	MbusCode    uint32
	FrameSizes  []v4l2.FrameSizeEnum
}

func getCapability(d *usbdevice.Device) (interface{}, error) {
	c := d.Capability()
	verVal := c.Version
	version := fmt.Sprintf("%d.%d.%d", verVal>>16, (verVal>>8)&0xff, verVal&0xff)
	var driverCapDescs []string
	for _, desc := range c.GetDriverCapDescriptions() {
		driverCapDescs = append(driverCapDescs, desc.Desc)
	}
	var deviceCapDescs []string
	for _, desc := range c.GetDeviceCapDescriptions() {
		deviceCapDescs = append(deviceCapDescs, desc.Desc)
	}

	capability := &Capability{
		Driver:     c.Driver,
		Card:       c.Card,
		BusInfo:    c.BusInfo,
		Version:    version,
		DriverCaps: driverCapDescs,
		DeviceCaps: deviceCapDescs,
	}
	return capability, nil
}

func getInputStatus(d *usbdevice.Device, index string) (uint32, error) {
	i, err := strconv.ParseUint(index, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("could not convert the given %s %s to Uint32", InputIndex, index)
	}
	info, err := d.GetVideoInputInfo(uint32(i))
	if err != nil {
		return 0, err
	}
	return info.GetStatus(), nil
}

func getDataFormat(d *usbdevice.Device) (interface{}, error) {
	pixFmt, err := d.GetPixFormat()
	if err != nil {
		return nil, err
	}
	pixDescription, _ := getPixFormatDesc(d, pixFmt.PixelFormat)
	dataFormat := DataFormat{
		Height:                 pixFmt.Height,
		Width:                  pixFmt.Width,
		PixelFormat:            pixFmt.PixelFormat,
		PixelFormatDescription: pixDescription,
		Field:                  v4l2.Fields[pixFmt.Field],
		BytesPerLine:           pixFmt.BytesPerLine,
		SizeImage:              pixFmt.SizeImage,
		Colorspace:             v4l2.Colorspaces[pixFmt.Colorspace],
	}

	xfunc := v4l2.XferFunctions[pixFmt.XferFunc]
	if pixFmt.XferFunc == v4l2.XferFuncDefault {
		xfunc = v4l2.XferFunctions[v4l2.ColorspaceToXferFunc(pixFmt.XferFunc)]
	}
	dataFormat.XferFunc = xfunc

	ycbcr := v4l2.YCbCrEncodings[pixFmt.YcbcrEnc]
	if pixFmt.YcbcrEnc == v4l2.YCbCrEncodingDefault {
		ycbcr = v4l2.YCbCrEncodings[v4l2.ColorspaceToYCbCrEnc(pixFmt.YcbcrEnc)]
	}
	dataFormat.YcbcrEnc = ycbcr

	quant := v4l2.Quantizations[pixFmt.Quantization]
	if pixFmt.Quantization == v4l2.QuantizationDefault {
		if v4l2.IsPixYUVEncoded(pixFmt.PixelFormat) {
			quant = v4l2.Quantizations[v4l2.QuantizationLimitedRange]
		} else {
			quant = v4l2.Quantizations[v4l2.QuantizationFullRange]
		}
	}
	dataFormat.Quantization = quant
	intervalCount := 0
	var frameRates []v4l2.Fract
	for {
		fd := d.Fd()
		index := uint32(intervalCount) // #nosec G115
		encoding := pixFmt.PixelFormat
		if interval, err := v4l2.GetFormatFrameInterval(fd, index, encoding, pixFmt.Width, pixFmt.Height); err == nil {
			intervalCount += 1
			// this swaps the internally tracked frame interval (seconds per frame)
			// to user-friendly frame rate (frames per second)
			frameRates = append(frameRates, v4l2.Fract{
				Denominator: interval.Interval.Max.Numerator,
				Numerator:   interval.Interval.Max.Denominator,
			})
		} else {
			break
		}
	}
	dataFormat.FrameRates = frameRates
	result := make(map[string]DataFormat)
	result[d.Name()] = dataFormat
	return result, nil
}

func getCropInfo(d *usbdevice.Device) (interface{}, error) {
	crop, err := d.GetCropCapability()
	if err != nil {
		return nil, err
	}
	result := make(map[string]v4l2.CropCapability)
	result[d.Name()] = crop
	return result, nil
}

func getStreamingParameters(d *usbdevice.Device) (interface{}, error) {
	sp, err := d.GetStreamParam()

	if err != nil {
		return nil, err
	}
	streamingParameters := StreamingParameters{}

	tpf := DescNotSpecified
	if sp.Capture.Capability == v4l2.StreamParamTimePerFrame {
		tpf = DescTimePerFrame
	}
	streamingParameters.Capability = StreamingCapability{tpf, sp.Capture.Capability}

	hiqual := DescNotSpecified
	if sp.Capture.CaptureMode == v4l2.StreamParamModeHighQuality {
		hiqual = DescHighQuality
	}
	streamingParameters.CaptureMode = CaptureMode{hiqual, sp.Capture.CaptureMode}

	streamingParameters.TimePerFrame = sp.Capture.TimePerFrame
	streamingParameters.ReadBuffers = sp.Capture.ReadBuffers

	result := make(map[string]StreamingParameters)
	result[d.Name()] = streamingParameters
	return result, nil
}

func getImageFormats(d *usbdevice.Device) (interface{}, error) {
	descs, err := d.GetFormatDescriptions()
	if err != nil {
		return nil, err
	}
	type result struct {
		ImageFormats []ImageFormat
	}
	var r result
	for _, desc := range descs {
		fss, err := v4l2.GetFormatFrameSizes(d.Fd(), desc.PixelFormat)
		if err != nil {
			return nil, err
		}
		pixDescription, _ := getPixFormatDesc(d, desc.PixelFormat)
		r.ImageFormats = append(r.ImageFormats, ImageFormat{
			Index:       desc.Index,
			BufType:     desc.StreamType,
			Flags:       desc.Flags,
			Description: pixDescription,
			PixelFormat: desc.PixelFormat,
			MbusCode:    desc.MBusCode,
			FrameSizes:  fss,
		})
	}
	resultMap := make(map[string]result)
	resultMap[d.Name()] = r

	return resultMap, nil
}

func getSupportedFrameRateFormats(d *usbdevice.Device) (interface{}, error) {
	descs, err := d.GetFormatDescriptions()
	if err != nil {
		return nil, err
	}

	type result struct {
		FrameRateFormats []FrameRateFormat
	}
	var r result
	for _, desc := range descs {
		var format FrameRateFormat
		format.Description, _ = getPixFormatDesc(d, desc.PixelFormat)
		fss, err := v4l2.GetFormatFrameSizes(d.Fd(), desc.PixelFormat)
		if err != nil {
			return nil, err
		}
		for _, frameSize := range fss {
			intervalCount := 0
			var frameInfo FrameInfo
			encoding := frameSize.PixelFormat
			height := frameSize.Size.MaxHeight
			width := frameSize.Size.MaxWidth
			frameInfo.FrameType = frameSize.Type
			frameInfo.Height = height
			frameInfo.Width = width
			frameInfo.PixelFormat = encoding
			frameInfo.Index = frameSize.Index
			for {
				fd := d.Fd()
				index := uint32(intervalCount) // #nosec G115
				if interval, err := v4l2.GetFormatFrameInterval(fd, index, encoding, width, height); err == nil {
					frameInfo.Rates = append(frameInfo.Rates, v4l2.Fract{
						// this swaps the internally tracked frame interval (seconds per frame)
						// to user-friendly frame rate (frames per second)
						Denominator: interval.Interval.Max.Numerator,
						Numerator:   interval.Interval.Max.Denominator,
					})
					intervalCount += 1
				} else {
					break
				}
			}
			format.FrameRates = append(format.FrameRates, frameInfo)
		}
		r.FrameRateFormats = append(r.FrameRateFormats, format)
	}
	resultMap := make(map[string]result)
	resultMap[d.Name()] = r
	return resultMap, nil
}

func GetFrameRate(d *usbdevice.Device) (interface{}, error) {
	streamParam, err := d.GetStreamParam()
	if err != nil {
		return nil, err
	}
	timePerFrame := streamParam.Capture.TimePerFrame
	var fps v4l2.Fract
	fps.Denominator = timePerFrame.Numerator
	fps.Numerator = timePerFrame.Denominator
	result := make(map[string]v4l2.Fract)
	result[d.Name()] = fps
	return result, nil
}

func getPixFormatDesc(usbDevice *usbdevice.Device, pixFmt uint32) (string, error) {
	descs, err := usbDevice.GetFormatDescriptions()
	if err != nil {
		return "", err
	}
	for _, desc := range descs {
		if pixFmt == desc.PixelFormat {
			return desc.Description, nil
		}
	}
	return "", nil
}
