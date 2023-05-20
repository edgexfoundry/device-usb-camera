// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v3/models"
	"os/exec"
	"strings"
)

const (
	maxStderrLines = 25
	ffmpegLogLevel = "info"
)

// runTranscoderWithOutput is based on transcoder.Transcoder.Run(), but tweaks a few things, and adds some
// quality of life improvements for the end user. It starts the transcoding process while also logging the ffmpeg
// output (from StdErr). StdErr text is also returned via the done error channel, so that it can be returned
// to the caller of a REST API. If an error occurs starting the process, it is returned immediately, and not
// via the error channel.
func (dev *Device) runTranscoderWithOutput() (<-chan error, error) {
	dev.mutex.Lock()
	defer dev.mutex.Unlock()

	t := dev.transcoder

	// generate the ffmpeg command line options, and prepend with some pre-defined options
	// -loglevel level+<ffmpegLogLevel>: will set the log level to ffmpegLogLevel and prefix output with the log level (for parsing)
	command := append([]string{"-loglevel", "level+" + ffmpegLogLevel}, t.GetCommand()...)
	if dev.lc.LogLevel() != models.TraceLog {
		// disable progress output if trace logging is not enabled
		command = append([]string{"-nostats"}, command...)
	}
	// -rtsp_transport tcp: force the rtsp transport to use tcp
	// these args must be put in the output section and not the first args, so just inject them right before the last
	// arg which is the rtsp url.
	command = append(command[0:len(command)-1], "-rtsp_transport", "tcp", command[len(command)-1])
	// todo: evaluate shell injection risks. if safe, use: // nolint: gosec
	proc := exec.Command(t.FFmpegExec(), command...)

	// Set the stdinPipe in case we need to stop the transcoding
	stdinPipe, err := proc.StdinPipe()
	if err != nil {
		dev.lc.Errorf("Ffmpeg Stdin not available: %s", err.Error())
	}

	var stdErrLines []string
	stdErrPipe, err := proc.StderrPipe()
	if err != nil {
		dev.lc.Errorf("Ffmpeg StderrPipe not available: %s. Unable to track output from process.", err.Error())
	} else {
		output := make(chan string, 10)
		// use a scanner to read the output of the pipe and send it to output channel
		go func() {
			defer close(output)
			scanner := bufio.NewScanner(stdErrPipe)
			scanner.Split(scanFFmpegLines)
			scanner.Buffer(make([]byte, 2), bufio.MaxScanTokenSize)

			for scanner.Scan() {
				// Scan the next line, redact it, and send it to output channel.
				output <- redact(scanner.Text())
			}
			dev.lc.Debugf("Output scanner complete for transcoder for device %s", dev.name)
		}()

		// keep track of stdErr text, so it can be returned to the caller via done channel
		go func() {
			for line := range output {
				// cap the size so that way the memory usage does not grow on commands with lots of output
				if len(stdErrLines) >= maxStderrLines {
					continue
				}

				line = strings.Trim(line, " ")
				if len(line) == 0 {
					continue // skip blank lines
				}

				// log the line at specific level depending on the content
				if strings.Contains(line, "[error]") || strings.Contains(line, "[fatal]") {
					stdErrLines = append(stdErrLines, line)
					dev.lc.Errorf("%s transcoder: %s", dev.name, line)
				} else if strings.Contains(line, "[warning]") {
					stdErrLines = append(stdErrLines, line)
					dev.lc.Warnf("%s transcoder: %s", dev.name, line)
				} else {
					// log everything else as debug, as ffmpeg info messages are just debug data to us
					dev.lc.Debugf("%s transcoder: %s", dev.name, line)
				}
			}
			dev.lc.Debugf("Done processing output for transcoder for device %s", dev.name)
		}()
	}

	// attempt to start the process
	if err = proc.Start(); err != nil {
		return nil, fmt.Errorf("failed to start FFMPEG transcoding for device %s (%s) with %s, message %s",
			dev.name, redact(strings.Join(command, " ")), err, strings.Join(stdErrLines, "\n"))
	}
	// only set the transcoder's process if we are successful in starting it
	t.SetProcess(proc)
	t.SetProcessStdinPipe(stdinPipe)
	dev.lc.Debugf("Set IsStreaming=true for device %s", dev.name)
	dev.streamingStatus.IsStreaming = true
	dev.streamingStatus.Error = ""

	dev.lc.Debugf("FFmpeg transcoder process for device %s has started with pid %d", dev.name, proc.Process.Pid)

	// in the background we will wait for the process to complete and return any errors over the done channel
	done := make(chan error)
	go func() {
		defer close(done)

		// wait until the process has exited
		err = proc.Wait()
		dev.lc.Debugf("FFmpeg process with pid %d for device %s exited with code %d. User time: %v, System time: %v",
			proc.Process.Pid, dev.name, proc.ProcessState.ExitCode(), proc.ProcessState.UserTime(), proc.ProcessState.UserTime())

		dev.mutex.Lock()
		dev.lc.Debugf("Set IsStreaming=false for device %s", dev.name)
		dev.streamingStatus.IsStreaming = false

		// if ffmpeg returned an error, add more details surrounding it
		if err != nil {
			err = fmt.Errorf("failed finish FFMPEG transcoding for device %s (%s) with %s message %s",
				dev.name, redact(strings.Join(command, " ")), err.Error(), strings.Join(stdErrLines, "\n"))
			dev.streamingStatus.Error = err.Error()
		} else {
			dev.streamingStatus.Error = ""
		}
		t.SetProcess(nil)
		t.SetProcessStdinPipe(nil)
		dev.mutex.Unlock()
		done <- err
	}()

	return done, nil
}

// scanFFmpegLines is based on bufio.ScanLines, however it will return a line as soon as
// it reaches a \r even if it is not followed by a \n. The reason for this is that ffmpeg
// sometimes uses \r as a way to replace the previous line, such as when progress is enabled.
// In those cases, the default bufio.ScanLines will miss those messages.
func scanFFmpegLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		// No more data. Return.
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, '\r'); i == 0 {
		return 1, nil, nil // Skip blank line.
	} else if i > 0 {
		// We have a cr terminated line
		return i + 1, data[0:i], nil
	}

	if i := bytes.IndexByte(data, '\n'); i == 0 {
		return 1, nil, nil // Skip blank line.
	} else if i > 0 {
		// We have a newline terminated line.
		return i + 1, data[0:i], nil
	}

	// If we are at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}

	// Request more data.
	return 0, nil, nil
}
