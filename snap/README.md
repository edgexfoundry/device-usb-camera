# EdgeX USB Camera Device Service Snap
[![edgex-device-usb-camera](https://snapcraft.io/edgex-device-usb-camera/badge.svg)](https://snapcraft.io/edgex-device-usb-camera)

This directory contains the snap packaging of the EdgeX USB Camera device service.

The snap is built automatically and published on the Snap Store as [edgex-device-usb-camera].

For usage instructions, please refer to Device Camera section in [Getting Started using Snaps][docs].

## Build from source
Execute the following command from the top-level directory of this repo:
```
snapcraft
```

This will create a snap package file with `.snap` extension. It can be installed locally by setting the `--dangerous` flag:
```bash
sudo snap install --dangerous <snap-file>
```

The [snapcraft overview](https://snapcraft.io/docs/snapcraft-overview) provides additional details.

### Obtain a Secret Store token
<!-- The `edgex-secretstore-token` snap slot makes it possible to automatically receive a token from a locally installed platform snap.-->

If the snap is built and installed locally, the interface will not auto-connect. You can check the status of the connections by running the `snap connections edgex-device-usb-camera` command.

To manually connect and obtain a token:
```bash
sudo snap connect edgexfoundry:edgex-secretstore-token edgex-device-usb-camera:edgex-secretstore-token
```

Please refer [here][secret-store-token] for further information.

Notes:
- The auto connection will not happen right now because the snap publisher isn't same as the `edgexfoundry` platrform snap (i.e. Canonical).
- The service isn't yet included in the secrets list of the platform snap.
Add the device service to list of services which have tokens generated for them:
```
# Enable app options
sudo snap set edgexfoundry app-options=true

# Additional secret store tokens
EXISTING=$(snap get edgexfoundry apps.security-secretstore-setup.config.add-secretstore-tokens)
sudo snap set edgexfoundry apps.security-secretstore-setup.config.add-secretstore-tokens="$EXISTING,device-usb-camera"

# Additional known secrets
EXISTING=$(snap get edgexfoundry apps.security-secretstore-setup.config.add-known-secrets)
sudo snap set edgexfoundry apps.security-secretstore-setup.config.add-known-secrets="$EXISTING,redisdb[device-usb-camera]"

# Additional registry ACL roles
EXISTING=$(snap get edgexfoundry apps.security-bootstrapper.config.add-registry-acl-roles)
sudo snap set edgexfoundry apps.security-bootstrapper.config.add-registry-acl-roles="$EXISTING,device-usb-camera"

# Run the bootstrappers:
sudo snap start edgexfoundry.security-secretstore-setup
sudo snap start edgexfoundry.security-consul-bootstrapper 

# Verify that the token has been generated and the service has been added to ADD_REGISTRY_ACL_ROLES environment variable by inspecting the following files:
# /var/snap/edgexfoundry/current/secrets/device-usb-camera/secrets-token.json
# /var/snap/edgexfoundry/current/config/security-bootstrapper/res/security-bootstrapper.env
```

### Connect the camera interface
The [`camera`](https://snapcraft.io/docs/camera-interface) interface is currently not automatically connected as it is pending auto-connection permissions from the store.
To connect manually:
```
snap connect edgex-device-usb-camera:camera :camera
```

[edgex-device-usb-camera]: https://snapcraft.io/edgex-device-usb-camera
[docs]: https://docs.edgexfoundry.org/2.2/getting-started/Ch-GettingStartedSnapUsers/#device-usb-camera
[secret-store-token]: https://docs.edgexfoundry.org/2.2/getting-started/Ch-GettingStartedSnapUsers/#secret-store-token

# Issues
The snap is 115 MB, exactly same as the docker image.

It may be possible to reduce the size by removing extra shared library object files.

See:
```
./device-usb-camera$ du -a -d 1 squashfs-root/usr/lib/x86_64-linux-gnu/ | sort -n -r | head -n 10
297884  squashfs-root/usr/lib/x86_64-linux-gnu/
89352   squashfs-root/usr/lib/x86_64-linux-gnu/libLLVM-12.so.1
43020   squashfs-root/usr/lib/x86_64-linux-gnu/dri
27392   squashfs-root/usr/lib/x86_64-linux-gnu/libicudata.so.66.1
15732   squashfs-root/usr/lib/x86_64-linux-gnu/libx265.so.179
14208   squashfs-root/usr/lib/x86_64-linux-gnu/libavcodec.so.58.54.100
14196   squashfs-root/usr/lib/x86_64-linux-gnu/libcodec2.so.0.9
9380    squashfs-root/usr/lib/x86_64-linux-gnu/librsvg-2.so.2.47.0
4808    squashfs-root/usr/lib/x86_64-linux-gnu/libflite_cmu_time_awb.so.2.1
4712    squashfs-root/usr/lib/x86_64-linux-gnu/libflite_cmu_us_rms.so.2.1
```
