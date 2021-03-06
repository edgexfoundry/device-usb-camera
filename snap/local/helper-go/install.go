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
	"os"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
	"github.com/canonical/edgex-snap-hooks/v2/env"
	"github.com/canonical/edgex-snap-hooks/v2/log"
)

// installProfiles copies the profile configuration.toml files from $SNAP to $SNAP_DATA.
func installConfig() error {
	resPath := "/config/device-usb-camera/res"
	err := os.MkdirAll(env.SnapData+resPath, 0755)
	if err != nil {
		return err
	}

	path := resPath + "/configuration.toml"
	err = hooks.CopyFile(
		env.Snap+path,
		env.SnapData+path)
	if err != nil {
		return err
	}

	return nil
}

func installDevices() error {
	devicesDir := "/config/device-usb-camera/res/devices"

	err := os.MkdirAll(env.SnapData+devicesDir, 0755)
	if err != nil {
		return err
	}

	err = hooks.CopyFile(
		hooks.Snap+devicesDir+"/general.usb.camera.toml.example",
		hooks.SnapData+devicesDir+"/general.usb.camera.toml.example")
	if err != nil {
		return err
	}

	return nil
}

func installDevProfiles() error {
	profilesDir := "/config/device-usb-camera/res/profiles"

	err := os.MkdirAll(env.SnapData+profilesDir, 0755)
	if err != nil {
		return err
	}

	err = hooks.CopyFile(
		hooks.Snap+profilesDir+"/general.usb.camera.yaml",
		hooks.SnapData+profilesDir+"/general.usb.camera.yaml")
	if err != nil {
		return err
	}

	return nil
}

func installRTSPConfig() error {
	confPath := "/config/rtsp-simple-server"
	err := os.MkdirAll(env.SnapData+confPath, 0755)
	if err != nil {
		return err
	}

	path := confPath + "/config.yml"
	err = hooks.CopyFile(
		env.Snap+path,
		env.SnapData+path)
	if err != nil {
		return err
	}

	return nil
}

func install() {
	log.SetComponentName("install")

	err := installConfig()
	if err != nil {
		log.Fatalf("error installing config file: %s", err)
	}

	err = installDevices()
	if err != nil {
		log.Fatalf("error installing devices config: %s", err)
	}

	err = installDevProfiles()
	if err != nil {
		log.Fatalf("error installing device profiles config: %s", err)
	}

	err = installRTSPConfig()
	if err != nil {
		log.Fatalf("error installing rtsp server config file: %s", err)
	}
}
