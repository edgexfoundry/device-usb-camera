# EdgeX USB Camera Device Service Snap
[![edgex-device-usb-camera](https://snapcraft.io/edgex-device-usb-camera/badge.svg)](https://snapcraft.io/edgex-device-usb-camera)

This directory contains the snap packaging of the EdgeX USB Camera device service.

The snap is built automatically and published on the Snap Store as [edgex-device-usb-camera].

For usage instructions, please refer to Device Camera section in [Getting Started using Snaps][docs].

## Build from source
Execute the following command from the top-level directory of this repo:
```
snapcraft -v
```

This will create a snap package file with `.snap` extension. It can be installed locally by setting the `--dangerous` flag:
```bash
sudo snap install --dangerous <snap-file>
```

The [snapcraft overview](https://snapcraft.io/docs/snapcraft-overview) provides additional details.

### Obtain a Secret Store token
The `edgex-secretstore-token` snap slot makes it possible to automatically receive a token from a locally installed platform snap.

If the snap is built and installed locally, the interface will not auto-connect. You can check the status of the connections by running the `snap connections edgex-device-usb-camera` command.

To manually connect and obtain a token:
```bash
sudo snap connect edgexfoundry:edgex-secretstore-token edgex-device-usb-camera:edgex-secretstore-token
```

Please refer [here][secret-store-token] for further information.

### Connect the camera interface
The [`camera`](https://snapcraft.io/docs/camera-interface) interface provides automatic access to all cameras.

However, if the snap is built and installed locally, the interface will not auto-connect. You can check the connection status by running the following command: `snap connections edgex-device-usb-camera`.

To manually connect:
```
snap connect edgex-device-usb-camera:camera :camera
```
Please refer [here][device-usb-camera] for further information.

[edgex-device-usb-camera]: https://snapcraft.io/edgex-device-usb-camera
[docs]: https://docs.edgexfoundry.org/3.0/getting-started/Ch-GettingStartedSnapUsers/#device-usb-camera
[secret-store-token]: https://docs.edgexfoundry.org/3.0/getting-started/Ch-GettingStartedSnapUsers/#secret-store-token
[device-usb-camera]: https://docs.edgexfoundry.org/3.0/getting-started/Ch-GettingStartedSnapUsers/#device-usb-camera
