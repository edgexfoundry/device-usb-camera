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
	"path/filepath"

	hooks "github.com/canonical/edgex-snap-hooks/v2"
	"github.com/canonical/edgex-snap-hooks/v2/env"
	"github.com/canonical/edgex-snap-hooks/v2/log"
)

// installProfiles copies the profile configuration.toml files from $SNAP to $SNAP_DATA.
func installConfig() error {
	var err error

	// serviceConfig := "/config/device-usb-camera/res/configuration.toml"
	// destFile := env.SnapData + res + "configuration.toml"
	// srcFile := env.Snap + path

	// output, err := exec.Command("ls", srcFile).CombinedOutput()
	// if err != nil {
	// 	return fmt.Errorf("%s: %s", err, output)
	// }
	// log.Errorf("%s", output)

	if err = os.MkdirAll(hooks.SnapData+"/config/device-usb-camera/res", 0755); err != nil {
		return err
	}

	path := "/config/device-usb-camera/res/configuration.toml"
	if err = hooks.CopyFile(
		env.Snap+path,
		env.SnapData+path); err != nil {
		return err
	}

	path = "/config/rtsp-simple-server.yml"
	if err = hooks.CopyFile(
		env.Snap+path,
		env.SnapData+path); err != nil {
		return err
	}

	return nil
}

// func installDevices() error {
// 	var err error

// 	path := "/config/device-camera/res/devices/camera.toml"
// 	destFile := hooks.SnapData + path
// 	srcFile := hooks.Snap + path

// 	if err = os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
// 		return err
// 	}

// 	if err = hooks.CopyFile(srcFile, destFile); err != nil {
// 		return err
// 	}

// 	return nil
// }

func installDevProfiles() error {
	var err error

	path := "/config/device-usb-camera/res/profiles/general.usb.camera.yaml"
	destFile := hooks.SnapData + path
	srcFile := hooks.Snap + path

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return err
	}

	if err = hooks.CopyFile(srcFile, destFile); err != nil {
		return err
	}

	return nil
}

func main() {
	log.SetComponentName("install")

	err := installConfig()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	// err = installDevices()
	// if err != nil {
	// 	log.Error(err)
	// 	os.Exit(1)
	// }

	err = installDevProfiles()
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
