// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"regexp"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/errors"
)

const (
	redactedStr = "//<redacted>@"
)

var (
	userPassRegex = regexp.MustCompile(`//(\S+):(\S+)@`)
)

type EdgeXErrorWrapper struct{}

func (e EdgeXErrorWrapper) CommandError(command string, err error) errors.EdgeX {
	return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to execute %s command", command), err)
}

// redact removes all instances of basic auth (ie. rtsp://username:password@server) from a url
func redact(val string) string {
	return userPassRegex.ReplaceAllString(val, redactedStr)
}
