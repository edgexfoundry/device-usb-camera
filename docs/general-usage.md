# General Usage

## Contents
[Configure and Build the service](#configure-and-build-the-service)  
[Run device-usb-service](#run-device-usb-camera)  
[Next Steps](#next-steps)


## Configure and Build the Service

### Define the device profile

Each device resource should have a mandatory attribute named `command` to indicate what action the device service should take for it.

Commands can be one of two types:

* Commands starting with `METADATA_` prefix are used to get camera metadata.

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
<p align="left">
      <i>Sample: Snippet from general.usb.camera.yaml</i>
</p>

* Commands starting with `VIDEO_` prefix are related to video stream.

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

<p align="left">
      <i>Sample: Snippet from general.usb.camera.yaml</i>
</p>

For all supported commands, refer to the sample at [general.usb.camera.yaml](../cmd/res/profiles/general.usb.camera.yaml).
> NOTE: In general, this sample should be applicable to all types of USB cameras.
> NOTE: You don't need to define device profile yourself unless you want to modify resource names or set default values for [video options](./advanced-options.md#video-options).

### Define the device

The device's protocol properties contain:
* `Path` is a file descriptor of camera created by OS. You can find the path of the connected USB camera through [v4l2-ctl](https://linuxtv.org/wiki/index.php/V4l-utils) utility.
* `AutoStreaming` indicates whether the device service should automatically start video streaming for cameras. Default value is false.

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
<p align="left">
      <i>Sample: Snippet from general.usb.camera.toml.example</i>
</p>

See the examples at `cmd/res/devices`  
> NOTE: When a new device is created in Core Metadata, a callback function of the device service will be called to add the device card name and serial number to protocol properties for identification purposes. These two pieces of information are obtained through `V4L2` API and `udev` utility.

### Build the executable file

```shell
make build
```

### Build docker image
```shell
make docker
```

## Run device-usb-camera
- Docker
  - For non secure mode
    ```
    make gen ds-usb-camera no-secty
    ```
  - For secure mode 
    ```
    make gen ds-usb-camera
    ```
  - Docker Compose start command
    ```
    docker-compose -p edgex up -d
    ```

  - Docker Compose clean command
    ```bash
    make clean
    ```  
- Native
  ```
  cd cmd && EDGEX_SECURITY_SECRET_STORE=false ./device-usb-camera
  ```

## Next Steps

[Explore more advanced options](./advanced-options.md)

## License
[Apache-2.0](../LICENSE)
