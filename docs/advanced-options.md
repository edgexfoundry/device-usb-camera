# Advanced Options

## Contents
[Video Options](#video-options)  
[Dynamic Discovery](#dynamic-discovery)  
[Camera Paths](#keep-the-paths-of-existing-camera-up-to-date) 
[RTSP Server](#configurable-rtsp-server-hostname-and-port)

## Video options
There are two types of options:
- The options start with `Input` prefix are used for the camera, such as specifying the image size and pixel format.
- The options start with `Output` prefix are used for the output video, such as specifying aspect ratio and quality.

These options can be passed in through object value when calling `StartStreaming`.

Query parameter:
- `device name`: The name of the camera

For example:
```shell
curl -X PUT -d '{
    "StartStreaming": {
      "InputImageSize": "640x480",
      "OutputVideoQuality": "5"
    }
}' http://localhost:59882/api/v2/device/name/<device name>/StartStreaming
```

Supported Input options:
- `InputFps`: Ignore original timestamps and instead generate timestamps assuming constant frame rate fps. (default - same as source)
- `InputImageSize`: Specifies the image size of the camera. The format is `wxh`, for example "640x480". (default - automatically selected by FFmpeg)
- `InputPixelFormat`: Set the preferred pixel format (for raw video). (default - automatically selected by FFmpeg)

Supported Output options:
- `OutputFrames`: Set the number of video frames to output. (default - no limitation on frames)
- `OutputFps`: Duplicate or drop input frames to achieve constant output frame rate fps. (default - same as InputFps)
- `OutputImageSize`: Performs image rescaling. The format is `wxh`, for example "640x480". (default - same as InputImageSize)
- `OutputAspect`: Set the video display aspect ratio specified by aspect. For example "4:3", "16:9". (default - same as source)
- `OutputVideoCodec`: Set the video codec. For example "mpeg4", "h264". (default - mpeg4)
- `OutputVideoQuality`: Use fixed video quality level. Range is a integer number between 1 to 31, with 31 being the worst quality. (default - dynamically set by FFmpeg)

You can also set default values for these options by adding additional attributes to the device resource `StartStreaming`.
The attribute name consists of a prefix "default" and the option name.

For example:
```yaml
deviceResources:
  - name: "StartStreaming"
   description: "Start streaming process."
   attributes:
     { command: "VIDEO_START_STREAMING",    
       defaultInputFrameSize: "320x240", 
       defaultOutputVideoQuality: "31" 
     }
   properties:
     valueType: "Object"
     readWrite: "W"
```

> NOTE: It's NOT recommended to set default video options in the [cmd/res/profiles/general.usb.camera.yaml](cmd/res/profiles/general.usb.camera.yaml) as they may not be supported by every camera.

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

### Rediscovery
The device service is able to rediscover and update devices that have been discovered previously.
Nothing additional is needed to enable this. It will run whenever the discover call is sent, regardless
of whether it is a manual or automated call to discover. The steps to configure discovery or to 
manually trigger discovery is explained [here](#enable-the-dynamic-discovery-function)

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

## Keep the paths of existing camera up to date
The paths (/dev/video*) of the connected cameras may change whenever the cameras are re-connected or the system restarts.
To ensure the paths of the existing cameras are up to date, the device service scans all the existing cameras to check whether their serial numbers match the connected cameras.
If there is a mismatch between them, the device service will scan all paths to find the matching device and update the existing device with the correct path.

This check can also be triggered by using the Device Service API `/refreshdevicepaths`.
For example:
```shell
curl -X POST http://localhost:59983/api/v2/refreshdevicepaths
```

It's recommended to trigger a check after re-plugging cameras.

## Configurable RTSP server hostname and port

The hostname and port of the RTSP server to which the device service publishes video streams can be configured in the [Driver] section of the service configuration located in the [configuration.toml](../cmd/res/configuration.toml).

For example:
```yaml
[Driver]
  RtspServerHostName = "localhost"
  RtspTcpPort = "8554"
```

## License
[Apache-2.0](LICENSE)
