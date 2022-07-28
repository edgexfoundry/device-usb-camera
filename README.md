# USB Camera Device Service
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-usb-camera/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-usb-camera/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-usb-camera/branch/main/graph/badge.svg?token=K4V4LAJYYW)](https://codecov.io/gh/edgexfoundry/device-usb-camera) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-usb-camera)](https://goreportcard.com/report/github.com/edgexfoundry/device-usb-camera) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-usb-camera?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-usb-camera/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-usb-camera?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-usb-camera)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-usb-camera) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-usb-camera)](https://github.com/edgexfoundry/device-usb-camera/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-usb-camera)](https://github.com/edgexfoundry/device-usb-camera/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-usb-camera-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-usb-camera)](https://github.com/edgexfoundry/device-usb-camera/commits)


## Overview
EdgeX device service for communicating with USB cameras attached to Linux OS platforms.
This service provides the following capabilities:
- Camera metadata
- Camera status
- Video stream reference

## Tested Devices
The following devices have been tested with EdgeX:
<!-- sorted alphabetically -->
- AUKEY PC-LM1E Webcam
- HP w200 Webcam
- Jinpei JW-01B USB FHD Web Computer Camera
- Logitech Brio 4K
- Logitech C270 HD Webcam
- Logitech StreamCam

## How does the device service work?
- The device service ONLY works on Linux with kernel v5.10 or higher.
- The device service uses V4L2 API to get camera metadata.
- The device service uses FFmpeg framework to capture video frames and stream them to an RTSP server.
- An [RTSP server](https://github.com/aler9/rtsp-simple-server) is embedded in the dockerized device service. 

## General Usage

### Build the executable file
```shell
make build
```

Note if your build fails with errors like:

```
.../go4vl@v0.0.2/v4l2/capability.go:48:33: could not determine kind of name for C.V4L2_CAP_IO_MC
.../go4vl@v0.0.2/v4l2/capability.go:46:33: could not determine kind of name for C.V4L2_CAP_META_OUTPUT
```

You are missing the appropriate kernal headers needed by the `github.com/vladimirvivien/go4vl` module
One possible solution is to manually download and install a more recent version of the libc-dev for your OS.

In the case of Ubuntu 20.04, one is not available in the normal repositories, so you can get it via these steps:

```
wget https://launchpad.net/~canonical-kernel-team/+archive/ubuntu/bootstrap/+build/20950478/+files/linux-libc-dev_5.10.0-14.15_amd64.deb
sudo dpkg -i linux-libc-dev_5.10.0-14.15_amd64.deb
```

### Build docker image

```shell
make docker
```

### Run device-usb-camera
- Docker
  - Use [compose-builder](https://github.com/edgexfoundry/edgex-compose/tree/main/compose-builder)
  - For non secure mode
    ```
    make gen ds-usb-camera no-secty
    ```
  - For secure mode 
    ```
    make gen ds-usb-camera
    ```
  - docker-compose -p edgex up -d

- Native
  ```
  cd cmd && EDGEX_SECURITY_SECRET_STORE=false ./device-usb-camera
  ```

### Define the device profile

Each device resource should have a mandatory attribute named `command` to indicate what action the device service should take for it.

There are two types of `command`:

* The commands started with **METADATA_** prefix are used to get camera metadata.

For example:
```yaml
deviceResources:
  - name: "CameraInfo"
    description: >-
      Camera information including driver name, device name, bus info, and capabilities.
      See https://www.kernel.org/doc/html/latest/userspace-api/media/v4l/vidioc-querycap.html.
    attributes:
      { command: "METADATA_DEVICE_CAPABILITY" }
    properties:
      valueType: "Object"
      readWrite: "R"
```

* The commands start with **VIDEO_** prefix are related to video stream.

For example:
```yaml
deviceResources:
  - name: "StreamURI"
    description: "Get video-streaming URI."
    attributes:
      { command: "VIDEO_STREAM_URI" }
    properties:
      valueType: "String"
      readWrite: "R"
```

For all supported commands, refer to the sample at [cmd/res/profiles/general.usb.camera.yaml](cmd/res/profiles/general.usb.camera.yaml).
> *Note: In general, this sample should be applicable to all types of USB cameras.
> You don't need to define device profile yourself unless you want to modify resource names or set default values for [video options](#Video options).

### Define the device

The device's protocol properties contain:
* **Path** is a file descriptor of camera created by OS. You can find the path of the connected USB camera through [v4l2-ctl](https://linuxtv.org/wiki/index.php/V4l-utils) utility.
* **AutoStreaming** indicates whether the device service should automatically start video streaming for cameras. Default value is false.

For example:
```yaml
[[DeviceList]]
  Name = "hp-w200-01"
  ProfileName = "USB-Camera-General"
  Description = "HP Webcam w200 - 01"
  Labels = [ "device-usb-camera-example" ]
  [DeviceList.Protocols]
    [DeviceList.Protocols.USB]
    Path = "/dev/video0"
    AutoStreaming = "false"
```
See the examples at [cmd/res/devices](cmd/res/devices)

> *Note: When a new device is created in Core Metadata, an callback function of the device service will be called to add the device card name and serial number to protocol properties for identification purposes.
> These two pieces of information are obtained through `V4L2` API and `udev` utility.

## Advanced topics

### Video options
There are two types of options:
- The options start with **Input** prefix are used for the camera, such as specifying the image size and pixel format.
- The options start with **Output** prefix are used for the output video, such as specifying aspect ratio and quality.

These options can be passed in through Object value when calling StartStreaming.

For example:
```shell
curl -X PUT -d '{
    "StartStreaming": {
      "InputImageSize": "640x480",
      "OutputVideoQuality": "5"
    }
}' http://localhost:59882/api/v2/device/name/hp-w200-01/StartStreaming
```

Supported Input options:
- **InputFps**: Ignore original timestamps and instead generate timestamps assuming constant frame rate fps. (default - same as source)
- **InputImageSize**: Specifies the image size of the camera. The format is `wxh`, for example "640x480". (default - automatically selected by FFmpeg)
- **InputPixelFormat**: Set the preferred pixel format (for raw video). (default - automatically selected by FFmpeg)

> *Note: If the given option value is not supported by the device, FFmpeg will automatically select the best one supported by the device.

Supported Output options:
- **OutputFrames**: Set the number of video frames to output. (default - no limitation on frames)
- **OutputFps**: Duplicate or drop input frames to achieve constant output frame rate fps. (default - same as InputFps)
- **OutputImageSize**: Performs image rescaling. The format is `wxh`, for example "640x480". (default - same as InputImageSize)
- **OutputAspect**: Set the video display aspect ratio specified by aspect. For example "4:3", "16:9". (default - same as source)
- **OutputVideoCodec**: Set the video codec. For example "mpeg4", "h264". (default - mpeg4)
- **OutputVideoQuality**: Use fixed video quality level. Range is a integer number between 1 to 31, with 31 being the worst quality. (default - dynamically set by FFmpeg)

You can also set default values for these options by adding additional attributes to the device resource **StartStreaming**.
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

> *Note: It's NOT recommended to set default video options in the [cmd/res/profiles/general.usb.camera.yaml](cmd/res/profiles/general.usb.camera.yaml) as they may not be supported by every camera.

### Dynamic Discovery
The device service supports [dynamic discovery](https://docs.edgexfoundry.org/2.1/microservices/device/Ch-DeviceServices/#dynamic-provisioning).
During dynamic discovery, the device service scans all connected USB devices and sends the discovered cameras to Core Metadata.
The device name of the camera discovered by the device service is comprised of Card Name and Serial Number, and the characters colon, space and dot will be replaced with underscores as they are invalid characters for device names in EdgeX.
Take the camera Logitech C270 as an example, its Card Name is "C270 HD WEBCAM" and the Serial Number is "B1CF0E50" hence the device name - "C270_HD_WEBCAM-B1CF0E50".

> *Note: Card Name and Serial number are used by the device service to uniquely identify a camera, although those cheaply mass-produced cameras may have the same serial number.

#### Enable the Dynamic Discovery function
Dynamic discovery is disabled by default to save computing resources.
If you want the device service to run the discovery periodically, enable it and set a desired interval.
The interval value must be a [Go duration](https://pkg.go.dev/time#ParseDuration).

[Option 1] Enable from the configuration.toml
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

#### Provision watcher example
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

### Keep the paths of existing camera up to date
The paths (/dev/video*) of the connected cameras may change whenever the cameras are re-connected or the system restarts.
To ensure the paths of the existing cameras are up to date, the device service scans all the existing cameras to check whether their serial numbers match the connected cameras.
If there is a mismatch between them, the device service will scan all paths to find the matching device and update the existing device with the correct path.

This check can also be triggered by using the Device Service API `/refreshdevicepaths`.
For example:
```shell
curl -X POST http://localhost:59983/api/v2/refreshdevicepaths
```

It's recommended to trigger a check after re-plugging cameras.

### Configurable RTSP server hostname and port

The hostname and port of the RTSP server to which the device service publishes video streams can be configured in the [Driver] section of the service configuration.

For example:
```yaml
[Driver]
  RtspServerHostName = "localhost"
  RtspTcpPort = "8554"
```


### CameraStatus Command
Use the following query to determine the status of the camera.

URL parameter:
- **DeviceName**: The name of the camera
- **InputIndex**: indicates the current index of the video input (if a camera only has one source for video, the index needs to be set to '0')

```
curl -X GET http://localhost:59882/api/v2/device/name/<DeviceName>/CameraStatus?InputIndex=0 | jq -r '"CameraStatus: " + (.event.readings[].value|tostring)'
```

   Example Output: 
   ```
    CameraStatus: 0
   ```

   **Response meanings**:
| Response   | Description |
| ---------- | ----------- |
| 0 | Ready |
| 1 | No Power |
| 2 | No Signal |
| 3 | No Color |  

## License
[Apache-2.0](LICENSE)
