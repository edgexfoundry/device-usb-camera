// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"strconv"

	"github.com/vladimirvivien/go4vl/v4l2"
	usbdevice "github.com/vladimirvivien/go4vl/v4l2/device"
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
	FrameSizes  []v4l2.FrameSize
}

func getCapability(d *usbdevice.Device) (interface{}, error) {
	c, err := d.GetCapability()
	if err != nil {
		return nil, err
	}
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
	fmt, err := d.GetPixFormat()
	if err != nil {
		return nil, err
	}

	result := DataFormat{}
	result.Height = fmt.Height
	result.Width = fmt.Width
	result.PixelFormat = v4l2.PixelFormats[fmt.PixelFormat]
	result.Field = v4l2.Fields[fmt.Field]
	result.BytesPerLine = fmt.BytesPerLine
	result.SizeImage = fmt.SizeImage
	result.Colorspace = v4l2.Colorspaces[fmt.Colorspace]

	xfunc := v4l2.XferFunctions[fmt.XferFunc]
	if fmt.XferFunc == v4l2.XferFuncDefault {
		xfunc = v4l2.XferFunctions[v4l2.ColorspaceToXferFunc(fmt.XferFunc)]
	}
	result.XferFunc = xfunc

	ycbcr := v4l2.YCbCrEncodings[fmt.YcbcrEnc]
	if fmt.YcbcrEnc == v4l2.YCbCrEncodingDefault {
		ycbcr = v4l2.YCbCrEncodings[v4l2.ColorspaceToYCbCrEnc(fmt.YcbcrEnc)]
	}
	result.YcbcrEnc = ycbcr

	quant := v4l2.Quantizations[fmt.Quantization]
	if fmt.Quantization == v4l2.QuantizationDefault {
		if v4l2.IsPixYUVEncoded(fmt.PixelFormat) {
			quant = v4l2.Quantizations[v4l2.QuantizationLimitedRange]
		} else {
			quant = v4l2.Quantizations[v4l2.QuantizationFullRange]
		}
	}
	result.Quantization = quant

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
	p, err := d.GetCaptureParam()
	if err != nil {
		return nil, err
	}
	result := StreamingParameters{}

	tpf := DescNotSpecified
	if p.Capability == v4l2.StreamParamTimePerFrame {
		tpf = DescTimePerFrame
	}
	result.Capability = StreamingCapability{tpf, p.Capability}

	hiqual := DescNotSpecified
	if p.CaptureMode == v4l2.StreamParamModeHighQuality {
		hiqual = DescHighQuality
	}
	result.CaptureMode = CaptureMode{hiqual, p.CaptureMode}

	result.TimePerFrame = p.TimePerFrame
	result.ReadBuffers = p.ReadBuffers
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
		fss, err := v4l2.GetFormatFrameSizes(d.GetFileDescriptor(), desc.PixelFormat)
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
