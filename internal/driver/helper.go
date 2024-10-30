// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"fmt"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"regexp"

	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
)

var userPassRegex = regexp.MustCompile(`//(\S+):(\S+)@`)

const redactedStr = "//<redacted>@"

type EdgeXErrorWrapper struct{}

func (e EdgeXErrorWrapper) CommandError(command string, err error) errors.EdgeX {
	return errors.NewCommonEdgeX(errors.KindServerError, fmt.Sprintf("failed to execute %s command", command), err)
}

// redact removes all instances of basic auth (i.e. rtsp://username:password@server) from a url
func redact(val string) string {
	return userPassRegex.ReplaceAllString(val, redactedStr)
}

// slicesAreEqual returns true if two []string slices contain the same elements in the same exact order. It will return
// false if any elements are not equal, and will return false if the elements are in different order.
func slicesAreEqual(lc logger.LoggingClient, a, b any) bool {
	if a == nil && b == nil {
		return true
	}

	aSlice, ok := a.([]string)
	if !ok {
		lc.Errorf("Expected argument 'a' to slicesAreEqual to be a slice of strings, but got type %T", a)
		return false
	}
	bSlice, ok := b.([]string)
	if !ok {
		lc.Errorf("Expected argument 'b' to slicesAreEqual to be a slice of strings, but got type %T", b)
		return false
	}

	if len(aSlice) != len(bSlice) {
		return false
	}
	for i, val := range aSlice {
		if val != bSlice[i] {
			return false
		}
	}
	return true
}
