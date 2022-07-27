# USB Camera Device Service RTSP Streaming Guide

## Contents

[System Requirements](#system-requirements)
[How It Works](#how-it-works)
[Configuration](#configuration)

[Dependencies](#dependencies)  
[Deploy the Service](#deploy-edgex-and-usb-device-camera-microservice)  
[Verify the Service](#verify-service-and-device-profiles)   
[Manage Devices](#manage-devices)   
[Shutting Down](#shutting-down)  
[License](#license)

## System Requirements

- Intel&#8482; Core&#174; processor
- Ubuntu 20.04.4 LTS
- USB-compliant Camera

**Time to Complete**

10-20 minutes

**Other Requirements**

You must have administrator (sudo) privileges to execute the user guide commands.

## How It Works
For an explanation of the architecture, see the [User Guide](UserGuide.md#how-it-works).

## Overview
EdgeX device service for communicating with USB cameras attached to Linux OS platforms.
This service provides the following capabilities:
- V4L2 API to get camera metadata.
- Camera status
- Video stream reference
- FFmpeg framework to capture video frames and stream them to an RTSP server.
- An [RTSP server](https://github.com/aler9/rtsp-simple-server) is embedded in the dockerized device service. 

## Tested Devices
The following devices have been tested with EdgeX:
<!-- sorted alphabetically -->
- AUKEY PC-LM1E Webcam
- HP w200 Webcam
- Jinpei JW-01B USB FHD Web Computer Camera
- Logitech Brio 4K
- Logitech C270 HD Webcam
- Logitech StreamCam

## Dependencies
The software has dependencies, including Git, Docker, Docker Compose, and assorted tools (e.g., curl). Follow the instructions below to install any dependency that is not already installed.  

### Install Git
Install Git from the official repository as documented on the [Git SCM](https://git-scm.com/download/linux) site.

1. Update installation repositories:
   ```bash
   sudo apt update
   ```

2. Add the Git repository:
   ```bash
   sudo add-apt-repository ppa:git-core/ppa -y
   ```

3. Install Git:
   ```bash
   sudo apt install git
   ```

### Install Docker
Install Docker from the official repository as documented on the [Docker](https://docs.docker.com/engine/install/ubuntu/) site.

### Verify Docker
To enable running Docker commands without the preface of sudo, add the user to the Docker group. Then run Docker with the `hello-world` test.

1. Create Docker group:
   ```bash
   sudo groupadd docker
   ```
   >NOTE: If the group already exists, `groupadd` outputs a message: **groupadd: group `docker` already exists**. This is OK.

2. Add User to group:
   ```bash
   sudo usermod -aG docker $USER
   ```

3. Refresh the group:
   ```bash
   newgrp docker 
   ```

4. To verify the Docker installation, run `hello-world`:

   ```bash
   docker run hello-world
   ```
   A **Hello from Docker!** greeting indicates successful installation.

   ```bash
   Unable to find image 'hello-world:latest' locally
   latest: Pulling from library/hello-world
   2db29710123e: Pull complete 
   Digest: sha256:10d7d58d5ebd2a652f4d93fdd86da8f265f5318c6a73cc5b6a9798ff6d2b2e67
   Status: Downloaded newer image for hello-world:latest

   Hello from Docker!
   This message shows that your installation appears to be working correctly.
   ...
   ```

### Install Docker Compose
Install Docker from the official repository as documented on the [Docker Compose](https://docs.docker.com/compose/install/#install-compose) site. See the Linux tab. 

1. Download current stable Docker Compose:
   ```bash
   sudo curl -L "https://github.com/docker/compose/releases/download/1.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
   ```
   >NOTE: When this guide was created, version 1.29.2 was current.

2. Set permissions:
   ```bash
   sudo chmod +x /usr/local/bin/docker-compose
   ```

###  Download EdgeX Compose
Clone the EdgeX compose repository

   ```bash
   git clone https://github.com/edgexfoundry/edgex-compose.git
   ```

### Install Tools
Install the build, media streaming, and parsing tools:

   ```bash
   sudo apt install build-essential vlc ffmpeg jq curl
   ```

   - The device service ONLY works on Linux with kernel v5.10 or higher.
NOTE: An additional package may be required to build the device-use-camera service
```
 sudo dpkg -i linux-libc-dev_5.10.0-14.15_amd64.deb
 ```

### Tool Descriptions
The table below lists command line tools this guide uses to help with EdgeX configuration and device setup.

| Tool        | Description | Note |
| ----------- | ----------- |----------- |
| **curl**     | Allows the user to connect to services such as EdgeX. |Use curl to get transfer information either to or from this service. In the tutorial, use `curl` to communicate with the EdgeX API. The call will return a JSON object.|
| **jq**   |Parses the JSON object returned from the `curl` requests. |The `jq` command includes parameters that are used to parse and format data. In this tutorial, the `jq` command has been configured to return and format appropriate data for each `curl` command that is piped into it. |
| **base64**   | Converts data into the Base64 format.| |

>Table 1: Command Line Tools

## Get the Source Code

Clone the device-usb-camera repository:

   ```bash
   git clone https://github.com/edgexfoundry/device-usb-camera.git
   ```
## Deploy EdgeX and USB Device Camera Microservice
### Building the docker image
```shell
make docker
```

## Configuration options
### Configurable RTSP server hostname and port
The hostname and port of the RTSP server to which the device service publishes video streams can be configured in the [Driver] section of the service configuration located at '/cmd/res/configuration.toml'. The default vaules can be used in this guide.

For example:
```yaml
[Driver]
  RtspServerHostName = "localhost"
  RtspTcpPort = "8554"
```
### Run the Service

1. Navigate to the `edgex-compose/compose-builder` directory.

2. Run EdgeX with the microservice:

   ```bash
    make run ds-usb-camera no-secty
   ```

## Verify Service and Device Profiles

1. Check the status of the container:

   ```bash 
   docker ps
   ```

   The status column will indicate if the container is running and how long it has been up.

   Example Output:

   ```docker
   CONTAINER ID   IMAGE                                         COMMAND                  CREATED       STATUS          PORTS                                                                                         NAMES
    f0a1c646f324   edgexfoundry/device-usb-camera:0.0.0-dev                        "/docker-entrypoint.…"   26 hours ago   Up 20 hours   127.0.0.1:8554->8554/tcp, 127.0.0.1:59983->59983/tcp                         edgex-device-usb-camera                                                                   edgex-device-onvif-camera
   ```

2. Check that the device service is added to EdgeX:

   ```bash
   curl -s http://localhost:59881/api/v2/deviceservice/name/device-usb-camera | jq
   ```
   Successful:
   ```json
  {
    "apiVersion": "v2",
    "statusCode": 200,
    "service": {
      "created": 1658769423192,
      "modified": 1658872893286,
      "id": "04470def-7b5b-4362-9958-bc5ff9f54f1e",
      "name": "device-usb-camera",
      "baseAddress": "http://edgex-device-usb-camera:59983",
      "adminState": "UNLOCKED"
    }
  }
   ```
   Unsuccessful:
   ```json
   {
      "apiVersion": "v2",
      "message": "fail to query device service by name device-usb-camera",
      "statusCode": 404
   }
   ```
## Adding Devices using REST API
Devices can either be added to the service by defining them in a static configuration file, discovering devices dynamically, or with the REST API. For this example, the device will be added using the REST API.

1. Edit the information to appropriately match the camera. The field `Path` should match that of the camera:

The device's protocol properties contain:
* **Path** is a file descriptor of camera created by OS. You can find the path of the connected USB camera through [v4l2-ctl](https://linuxtv.org/wiki/index.php/V4l-utils) utility.
* **AutoStreaming** indicates whether the device service should automatically start video streaming for cameras. Default value is false.

   ```bash
   curl -X POST -H 'Content-Type: application/json'  \
   http://localhost:59881/api/v2/device \
   -d '[
            {
               "apiVersion": "v2",
               "device": {
                  "name":"Camera001",
                  "serviceName": "device-usb-camera",
                  "profileName": "USB-Camera-General",
                  "description": "My test camera",
                  "adminState": "UNLOCKED",
                  "operatingState": "UP",
                  "protocols": {
                	  "USB": {
                    	"CardName": "NexiGo N930AF FHD Webcam: NexiG",
                    	"Path": "/dev/video6",
 			                "AutoStreaming": "false"
                }
            }
               }
            }
   ]'
   ```

   Example Output: 
   ```bash
   [{"apiVersion":"v2","statusCode":201,"id":"fb5fb7f2-768b-4298-a916-d4779523c6b5"}]
   ```

### Start Video Streaming
Unless the device service is configured to stream video from the camera automatically, a 'StartStreaming' command must be sent to the device service.

There are two types of options:
- The options start with **Input** prefix are used for the camera, such as specifying the image size and pixel format.
- The options start with **Output** prefix are used for the output video, such as specifying aspect ratio and quality.

These options can be passed in through Object value when calling StartStreaming.

Query parameter:
- **DeviceName**: The name of the camera

For example:
```shell
curl -X PUT -d '{
    "StartStreaming": {
      "InputImageSize": "640x480",
      "OutputVideoQuality": "5"
    }
}' http://localhost:59882/api/v2/device/name/[DeviceName]/StartStreaming
```

Supported Input options:
- **InputFps**: Ignore original timestamps and instead generate timestamps assuming constant frame rate fps. (default - same as source)
- **InputImageSize**: Specifies the image size of the camera. The format is `wxh`, for example "640x480". (default - automatically selected by FFmpeg)
- **InputPixelFormat**: Set the preferred pixel format (for raw video). (default - automatically selected by FFmpeg)

Supported Output options:
- **OutputFrames**: Set the number of video frames to output. (default - no limitation on frames)
- **OutputFps**: Duplicate or drop input frames to achieve constant output frame rate fps. (default - same as InputFps)
- **OutputImageSize**: Performs image rescaling. The format is `wxh`, for example "640x480". (default - same as InputImageSize)
- **OutputAspect**: Set the video display aspect ratio specified by aspect. For example "4:3", "16:9". (default - same as source)
- **OutputVideoCodec**: Set the video codec. For example "mpeg4", "h264". (default - mpeg4)
- **OutputVideoQuality**: Use fixed video quality level. Range is a integer number between 1 to 31, with 31 being the worst quality. (default - dynamically set by FFmpeg)


### Determine Stream Uri of Camera
The device service provides a way to determine the stream URI of a camera.

Query parameter:
- **DeviceName**: The name of the camera

```
curl -s http://localhost:59882/api/v2/device/name/[DeviceName]/StreamURI | jq -r '"StreamURI: " + '.event.readings[].value''
```

The response to the above call should look similar to the following:

```
StreamURI: rtsp://localhost:8554/stream/NexiGo_N930AF_FHD_Webcam__NexiG-20201217010
```

### Stream the RTSP stream. 

   ffplay can be used to stream. The command follows this format: 
   
   `ffplay -rtsp_transport tcp rtsp://<IP address>:<port>/<streamname>`.

   Using the `streamURI` returned from the previous step, run ffplay:
   
   ```bash
   ffplay -rtsp_transport tcp rtsp://192.168.86.34:8554/stream1
   ```

  - To shut down ffplay, use the ctrl-c command.
## Shutting Down
To stop all EdgeX services (containers), execute the `make down` command:

1. Navigate to the `edgex-compose/compose-builder` directory.
1. Run this command
   ```bash
   make down
   ```
1. To shut down and delete all volumes, run this command
   ```bash
   make clean
   ```

## Troubleshooting
### StreamingStatus
To verify the usb camera is set to stream video, use the command below. 

```
curl http://localhost:59882/api/v2/device/name/[DeviceName]/StreamingStatus | jq -r '"StreamingStatus: " + (.event.readings[].objectValue.IsStreaming|tostring)'
```
- please replace [DeviceName] with the name of the device you want to test
- if the StreamingStatus is false, the camera is not configured to stream video. Please try the Start Video Streaming section again

## License
[Apache-2.0](LICENSE)
