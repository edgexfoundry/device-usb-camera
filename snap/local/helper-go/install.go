/*
 * Copyright (C) 2022 Canonical Ltd
 *
 *  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 * SPDX-License-Identifier: Apache-2.0'
 */

package main

import (
	hooks "github.com/canonical/edgex-snap-hooks/v3"
	"github.com/canonical/edgex-snap-hooks/v3/env"
	"github.com/canonical/edgex-snap-hooks/v3/log"
)

func installUSBCameraConfig() error {
	path := "/config/" + usbCameraApp + "/res"

	return hooks.CopyDir(env.Snap+path, env.SnapData+path)
}

func installRTSPConfig() error {
	path := "/config/" + rtspServerApp

	return hooks.CopyDir(env.Snap+path, env.SnapData+path)
}

func install() {
	log.SetComponentName("install")

	if err := installUSBCameraConfig(); err != nil {
		log.Fatalf("Error installing Device USB Camera files: %s", err)
	}

	if err := installRTSPConfig(); err != nil {
		log.Fatalf("Error installing RTSP Simple Server config file: %s", err)
	}
}
