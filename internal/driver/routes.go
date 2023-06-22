// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2022-2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package driver

import (
	"net/http"

	"github.com/edgexfoundry/go-mod-core-contracts/v3/common"
)

func (d *Driver) RefreshExistingDevicesPathsRoute(writer http.ResponseWriter, request *http.Request) {
	go d.RefreshExistingDevicesPaths()
	correlationID := request.Header.Get(common.CorrelationHeader)
	writer.Header().Set(common.CorrelationHeader, correlationID)
	writer.Header().Set(common.ContentType, common.ContentTypeJSON)
	writer.WriteHeader(http.StatusAccepted)
}
