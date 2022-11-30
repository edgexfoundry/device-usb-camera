# Dynamic Discovery
The device service supports [dynamic discovery](https://docs.edgexfoundry.org/2.1/microservices/device/Ch-DeviceServices/#dynamic-provisioning).
During dynamic discovery, the device service scans all connected USB devices and sends the discovered cameras to Core Metadata.
The device name of the camera discovered by the device service is comprised of Card Name and Serial Number, and the characters colon, space and dot will be replaced with underscores as they are invalid characters for device names in EdgeX.
Take the camera Logitech C270 as an example, it's Card Name is "C270 HD WEBCAM" and the Serial Number is "B1CF0E50" hence the device name - "C270_HD_WEBCAM-B1CF0E50".

> NOTE: Card Name and Serial number are used by the device service to uniquely identify a camera. Some manufactures, however, may not support unique serial numbers for their cameras. Please check with your camera manufacturer.

## Dynamic Discovery function
Dynamic discovery is enabled by default to make setup easier. It can be disable by changing the `Enabled` option to `false` as shown below.

### Disable discovery

[Option 1] Disable discovery from the [configuration.toml](../cmd/res/configuration.toml) file.
```yaml
[Device] 
...
    [Device.Discovery]
    Enabled = false
    Interval = "1h"
```

[Option 2] Disable discovery from environment variables. 
```shell
export DEVICE_DISCOVERY_ENABLED=false
export DEVICE_DISCOVERY_INTERVAL=1h
```

[Option 3] Disable discovery from a curl command.

//TODO: add discovery command
To stop the usb camera from live streaming, use the following command:

Query parameter:
- `device name`: The name of the camera

For example:
```shell
curl -X PUT -d '{
    "StopStreaming": "true"
}' http://localhost:59882/api/v2/device/name/<device name>/Discovery
```

### Enable discovery  
[Option 1] Enable discovery from the [configuration.toml](../cmd/res/configuration.toml) file.
```yaml
[Device] 
...
    [Device.Discovery]
    Enabled = true
    Interval = "1h"
```

[Option 2] Enable discovery from environment variables.
```shell
export DEVICE_DISCOVERY_ENABLED=true
export DEVICE_DISCOVERY_INTERVAL=1h
```


To manually trigger a Dynamic Discovery, use this [device service API](https://app.swaggerhub.com/apis-docs/EdgeXFoundry1/device-sdk/2.2.0#/default/post_discovery).  
The interval value must be a [Go duration](https://pkg.go.dev/time#ParseDuration).  

## Configure the Provision Watchers

> NOTE: This section is for manually adding provision watchers, one is already added by default.

The provision watcher sets up parameters for EdgeX to automatically add devices to core-metadata. They can be configured to look for certain features, as well as block features. The default provision watcher is sufficient unless you plan on having multiple different cameras with different profiles and resources. Learn more about provision watchers [here](https://docs.edgexfoundry.org/2.2/microservices/core/metadata/Ch-Metadata/#provision-watcher). The provision watchers can be edited manually in the folder located at `/cmd/res/provision_watchers`


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