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
	"strings"

	"github.com/canonical/edgex-snap-hooks/v2/log"
	"github.com/canonical/edgex-snap-hooks/v2/options"
	"github.com/canonical/edgex-snap-hooks/v2/snapctl"
)

// var cli *hooks.CtlCli = hooks.NewSnapCtl()

func main() {
	// var debug = false
	// // var err error
	// var envJSON string

	// status, err := cli.Config("debug")
	// if err != nil {
	// 	fmt.Println(fmt.Sprintf("edgex-device-camera:configure: can't read value of 'debug': %v", err))
	// 	os.Exit(1)
	// }
	// if status == "true" {
	// 	debug = true
	// }

	// if err = hooks.Init(debug, "edgex-device-camera"); err != nil {
	// 	fmt.Println(fmt.Sprintf("edgex-device-camera:configure: initialization failure: %v", err))
	// 	os.Exit(1)

	// }

	log.SetComponentName("configure")
	err := options.ProcessAppConfig("device-usb-camera")
	if err != nil {
		log.Errorf("could not process options: %v", err)
		os.Exit(1)
	}

	// cli := hooks.NewSnapCtl()
	// envJSON, err = cli.Config(hooks.EnvConfig)
	// if err != nil {
	// 	log.Errorf("Reading config 'env' failed: %v", err)
	// 	os.Exit(1)
	// }

	// if envJSON != "" {
	// 	log.Debugf("envJSON: %s", envJSON)
	// 	err = hooks.HandleEdgeXConfig("device-camera", envJSON, local.ConfToEnv)
	// 	if err != nil {
	// 		log.Errorf("HandleEdgeXConfig failed: %v", err)
	// 		os.Exit(1)
	// 	}
	// }

	// If autostart is not explicitly set, default to "no"
	// as only example service configuration and profiles
	// are provided by default.
	autostart, err := snapctl.Get("autostart").Run()
	if err != nil {
		log.Errorf("Reading config 'autostart' failed: %v", err)
		os.Exit(1)
	}
	if autostart == "" {
		log.Debug("autostart is NOT set, initializing to 'no'")
		autostart = "no"
	}
	autostart = strings.ToLower(autostart)
	log.Debugf("autostart=%s", autostart)

	// service is stopped/disabled by default in the install hook
	switch autostart {
	case "true", "yes":
		err = snapctl.Start("device-usb-camera").Enable().Run()
		if err != nil {
			log.Errorf("Can't start service: %s", err)
			os.Exit(1)
		}
	case "false", "no":
		// no action necessary
	default:
		log.Errorf("Invalid value for 'autostart': %s", autostart)
		os.Exit(1)
	}
}
