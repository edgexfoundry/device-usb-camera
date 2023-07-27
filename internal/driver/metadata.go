// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
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
	Width        uint32
	Height       uint32
	PixelFormat  string
	Field        string
	BytesPerLine uint32
	SizeImage    uint32
	Colorspace   string
	XferFunc     string
	YcbcrEnc     string
	Quantization string
	FrameRates   []v4l2.Fract
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
	PixelFormat string
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
	i, err := strconv.ParseUint(fmt.Sprintf("%v", index), 10, 32)
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

	result := DataFormat{}
	result.Height = pixFmt.Height
	result.Width = pixFmt.Width
	result.PixelFormat = v4l2.PixelFormats[pixFmt.PixelFormat]
	result.Field = v4l2.Fields[pixFmt.Field]
	result.BytesPerLine = pixFmt.BytesPerLine
	result.SizeImage = pixFmt.SizeImage
	result.Colorspace = v4l2.Colorspaces[pixFmt.Colorspace]

	xfunc := v4l2.XferFunctions[pixFmt.XferFunc]
	if pixFmt.XferFunc == v4l2.XferFuncDefault {
		xfunc = v4l2.XferFunctions[v4l2.ColorspaceToXferFunc(pixFmt.XferFunc)]
	}
	result.XferFunc = xfunc

	ycbcr := v4l2.YCbCrEncodings[pixFmt.YcbcrEnc]
	if pixFmt.YcbcrEnc == v4l2.YCbCrEncodingDefault {
		ycbcr = v4l2.YCbCrEncodings[v4l2.ColorspaceToYCbCrEnc(pixFmt.YcbcrEnc)]
	}
	result.YcbcrEnc = ycbcr

	quant := v4l2.Quantizations[pixFmt.Quantization]
	if pixFmt.Quantization == v4l2.QuantizationDefault {
		if v4l2.IsPixYUVEncoded(pixFmt.PixelFormat) {
			quant = v4l2.Quantizations[v4l2.QuantizationLimitedRange]
		} else {
			quant = v4l2.Quantizations[v4l2.QuantizationFullRange]
		}
	}
	result.Quantization = quant
	intervalCount := 0
	var frameRates []v4l2.Fract
	for {
		fd := d.Fd()
		index := uint32(intervalCount)
		encoding := pixFmt.PixelFormat
		if interval, err := v4l2.GetFormatFrameInterval(fd, index, encoding, pixFmt.Width, pixFmt.Height); err == nil {
			intervalCount += 1
			// this swaps the internally track frame interval (seconds per frame)
			// to user-friendly frame rate (frames per second)
			frameRates = append(frameRates, v4l2.Fract{
				Denominator: interval.Interval.Max.Numerator,
				Numerator:   interval.Interval.Max.Denominator,
			})
		} else {
			break
		}
	}
	result.FrameRates = frameRates

	return result, nil
}

func getCropInfo(d *usbdevice.Device) (interface{}, error) {
	crop, err := d.GetCropCapability()
	if err != nil {
		return nil, err
	}
	return crop, nil
}

func getStreamingParameters(d *usbdevice.Device) (interface{}, error) {
	sp, err := d.GetStreamParam()

	if err != nil {
		return nil, err
	}
	result := StreamingParameters{}

	tpf := DescNotSpecified
	if sp.Capture.Capability == v4l2.StreamParamTimePerFrame {
		tpf = DescTimePerFrame
	}
	result.Capability = StreamingCapability{tpf, sp.Capture.Capability}

	hiqual := DescNotSpecified
	if sp.Capture.CaptureMode == v4l2.StreamParamModeHighQuality {
		hiqual = DescHighQuality
	}
	result.CaptureMode = CaptureMode{hiqual, sp.Capture.CaptureMode}

	result.TimePerFrame = sp.Capture.TimePerFrame
	result.ReadBuffers = sp.Capture.ReadBuffers
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
		r.ImageFormats = append(r.ImageFormats, ImageFormat{
			Index:       desc.Index,
			BufType:     desc.StreamType,
			Flags:       desc.Flags,
			Description: desc.Description,
			PixelFormat: v4l2.PixelFormats[desc.PixelFormat],
			MbusCode:    desc.MBusCode,
			FrameSizes:  fss,
		})
	}
	return r, nil
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
		format.Description = desc.String()
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
				index := uint32(intervalCount)
				if interval, err := v4l2.GetFormatFrameInterval(fd, index, encoding, width, height); err == nil {
					frameInfo.Rates = append(frameInfo.Rates, v4l2.Fract{
						// this swaps the internally track frame interval (seconds per frame)
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
	return r, nil
}
