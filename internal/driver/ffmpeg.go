package driver

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/errors"

	"github.com/vladimirvivien/go4vl/v4l2"
)

type FFmpeg struct {
	inputOptions  []string
	outputOptions []string
}

func (f FFmpeg) ObtainOutputFrames(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegFrames, value}
	}
	return nil
}

func (f FFmpeg) ObtainInputFps(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegFps, value}
	}
	return nil
}

func (f FFmpeg) ObtainOutputFps(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegFps, value}
	}
	return nil
}

func (f FFmpeg) ObtainInputImageSize(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegSize, value}
	}
	return nil
}

func (f FFmpeg) ObtainOutputImageSize(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegSize, value}
	}
	return nil
}

func (f FFmpeg) ObtainOutputAspect(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegAspect, value}
	}
	return nil
}

func (f FFmpeg) ObtainOutputVideoQuality(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegQScale, value}
	}
	return nil
}

func (f FFmpeg) ObtainOutputVideoCodec(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegVCodec, value}
	}
	return nil
}

func (f FFmpeg) ObtainInputPixelFormat(value string) []string {
	if len(value) != 0 {
		return []string{FFmpegInputFormat, value}
	}
	return nil
}

func (f *FFmpeg) setOptions(name, val string) bool {
	opt := reflect.ValueOf(f).MethodByName(fmt.Sprintf("Obtain%s", name))
	if (opt != reflect.Value{}) {
		result := opt.Call([]reflect.Value{reflect.ValueOf(val)})
		if val, ok := result[0].Interface().([]string); ok {
			if strings.HasPrefix(name, PrefixInput) {
				f.inputOptions = append(f.inputOptions, val...)
			} else if strings.HasPrefix(name, PrefixOutput) {
				f.outputOptions = append(f.outputOptions, val...)
			} else {
				return false
			}
			return true
		}
	}
	return false
}

func setupFFmpegOptions(dev *Device, opts interface{}, attr map[string]interface{}) errors.EdgeX {
	options, ok := opts.(map[string]interface{})
	if !ok {
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			"failed to parse request body", nil)
	}

	ffmpeg := &FFmpeg{}
	// obtain FFmpeg options defined in request body
	for optName, value := range options {
		optVal, err := parseOptionValue(optName, value)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid,
				"failed to parse option value", err)
		}
		if ffmpeg.setOptions(optName, optVal) {
			dev.updateFFmpegOptions(optName, optVal)
			continue
		}
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("unsupported option: %s", optName), nil)
	}

	// obtain default FFmpeg options defined in resource attributes
	for name, value := range attr {
		if name == Command {
			continue
		}
		optName := strings.ReplaceAll(name, "default", "")
		if _, ok := options[optName]; ok {
			continue
		}
		optVal, err := parseOptionValue(optName, value)
		if err != nil {
			return errors.NewCommonEdgeX(errors.KindContractInvalid,
				"failed to parse option value", err)
		}
		if ffmpeg.setOptions(optName, optVal) {
			dev.updateFFmpegOptions(optName, optVal)
			continue
		}
		return errors.NewCommonEdgeX(errors.KindContractInvalid,
			fmt.Sprintf("unsupported option: %s", optName), nil)
	}

	if len(ffmpeg.inputOptions) > 0 {
		dev.transcoder.MediaFile().SetRawInputArgs(ffmpeg.inputOptions)
	}
	if len(ffmpeg.outputOptions) > 0 {
		dev.transcoder.MediaFile().SetRawOutputArgs(ffmpeg.outputOptions)
	}
	return nil
}

func parseOptionValue(name string, value interface{}) (string, error) {
	stringValue, ok := value.(string)
	if !ok {
		return stringValue, errors.NewCommonEdgeX(errors.KindContractInvalid,
			"value should be a string", nil)
	}

	if name == InputPixelFormat {
		switch value {
		case v4l2.PixelFormats[v4l2.PixFmtRGB24], FFmpegPixFmtRGB24:
			return FFmpegPixFmtRGB24, nil
		case v4l2.PixelFormats[v4l2.PixFmtGrey], FFmpegPixFmtGray:
			return FFmpegPixFmtGray, nil
		case v4l2.PixelFormats[v4l2.PixelFmtYUYV], FFmpegPixelFmtYUYV:
			return FFmpegPixelFmtYUYV, nil
		case v4l2.PixelFormats[v4l2.PixelFmtMJPEG], FFmpegPixelFmtMJPEG:
			// mjpeg is not in the list of available FFmpeg pixel formats, but it does work.
			return FFmpegPixelFmtMJPEG, nil
		default:
			// No corresponding pixel formats of FFmpeg for the following v4l2.PixelFormats:
			// v4l2.PixelFmtJPEG, v4l2.PixelFmtMPEG, v4l2.PixelFmtH264, and v4l2.PixelFmtMPEG4
			// For a full list of available FFmpeg pixel formats, use this command "ffmpeg -pix_fmts" with FFmpeg command-line tool
			return stringValue, fmt.Errorf(`invalid value "%s" for %s option`, value, name)
		}
	}
	return stringValue, nil
}
