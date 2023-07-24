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
	FpsIntervals []v4l2.Fract
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
	intervalCount := 0
	var intervals []v4l2.Fract
	for {
		fd := d.Fd()
		index := uint32(intervalCount)
		encoding := fmt.PixelFormat
		if interval, exit := v4l2.GetFormatFrameInterval(fd, index, encoding, fmt.Width, fmt.Height); exit == nil {
			intervalCount += 1
			intervals = append(intervals, interval.Interval.Max)
		} else {
			break
		}
	}
	result.FpsIntervals = intervals

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

func getSupportedIntervalFormats(d *usbdevice.Device) (interface{}, error) {
	descs, err := d.GetFormatDescriptions()
	if err != nil {
		return nil, err
	}
	type result struct {
		IntervalFormats [][][]v4l2.FrameIntervalEnum
	}
	var r result
	formats := make([][][]v4l2.FrameIntervalEnum, len(descs))
	for i, desc := range descs {
		fss, err := v4l2.GetFormatFrameSizes(d.Fd(), desc.PixelFormat)
		if err != nil {
			return nil, err
		}
		frameIntervals := make([][]v4l2.FrameIntervalEnum, len(fss))
		for j, frameSize := range fss {
			intervalCount := 0
			var intervals []v4l2.FrameIntervalEnum
			for {
				fd := d.Fd()
				index := uint32(intervalCount)
				encoding := frameSize.PixelFormat
				height := frameSize.Size.MaxHeight
				width := frameSize.Size.MaxWidth
				if interval, exit := v4l2.GetFormatFrameInterval(fd, index, encoding, width, height); exit == nil {
					intervals = append(intervals, interval)
					intervalCount += 1
				} else {
					break
				}
			}
			frameIntervals[j] = intervals
		}
		formats[i] = frameIntervals
	}
	r.IntervalFormats = formats
	return r, nil
}
