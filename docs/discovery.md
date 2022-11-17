
## Dynamic Discovery
The device service supports [dynamic discovery](https://docs.edgexfoundry.org/2.1/microservices/device/Ch-DeviceServices/#dynamic-provisioning).
During dynamic discovery, the device service scans all connected USB devices and sends the discovered cameras to Core Metadata.
The device name of the camera discovered by the device service is comprised of Card Name and Serial Number, and the characters colon, space and dot will be replaced with underscores as they are invalid characters for device names in EdgeX.
Take the camera Logitech C270 as an example, it's Card Name is "C270 HD WEBCAM" and the Serial Number is "B1CF0E50" hence the device name - "C270_HD_WEBCAM-B1CF0E50".

> NOTE: Card Name and Serial number are used by the device service to uniquely identify a camera. Some manufactures, however, may not support unique serial numbers for their cameras. Please check with your camera manufacturer.

### Enable the Dynamic Discovery function
Dynamic discovery is disabled by default to save computing resources.
If you want the device service to run the discovery periodically, enable it and set a desired interval.
The interval value must be a [Go duration](https://pkg.go.dev/time#ParseDuration).

[Option 1] Enable from the [configuration.toml](../cmd/res/configuration.toml)
```yaml
[Device] 
...
    [Device.Discovery]
    Enabled = true
    Interval = "1h"
```

[Option 2] Enable from the env
```shell
export DEVICE_DISCOVERY_ENABLED=true
export DEVICE_DISCOVERY_INTERVAL=1h
```

To manually trigger a Dynamic Discovery, use this [device service API](https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/device-sdk/2.2.0#/default/post_discovery).

### Configure the Provision Watchers

The provision watcher sets up parameters for EdgeX to automatically add devices to core-metadata. They can be configured to look for certain features, as well as block features. The default provision watcher is sufficient unless you plan on having multiple different cameras with different profiles and resources. Learn more about provision watchers [here](https://docs.edgexfoundry.org/2.2/microservices/core/metadata/Ch-Metadata/#provision-watcher).


```shell
curl -X POST \
-d '[
   {
      "provisionwatcher":{
         "apiVersion":"v2",
         "name":"USB-Camera-Provision-Watcher",
         "adminState":"UNLOCKED",
         "identifiers":{
            "Path": "."
         },
         "serviceName": "device-usb-camera",
         "profileName": "USB-Camera-General"
      },
      "apiVersion":"v2"
   }
]' http://localhost:59881/api/v2/provisionwatcher
```