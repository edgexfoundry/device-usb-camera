# Setup

## Contents
[System Requirements](#system-requirements)  
[Dependencies](#dependencies)  
&nbsp;&nbsp;&nbsp;&nbsp;[Git](#install-git)  
&nbsp;&nbsp;&nbsp;&nbsp;[Docker](#install-docker)   
&nbsp;&nbsp;&nbsp;&nbsp;[Tools](#install-tools)      
[Download EdgeX Compose](#download-edgex-compose)  
[Additional Installs](#additional-installs)  
[Next Steps](#next-steps)  

## System Requirements

- Intel&#8482; Core&#174; processor
- Ubuntu 20.04.4 LTS
- USB-compliant Camera

## Dependencies

The software has dependencies, including Git, Docker, Docker Compose, and assorted command line tools. Follow the instructions below to install any dependency that is not already installed.  

## Install Git
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

## Install Docker
Install Docker from the official repository as documented on the [Docker](https://docs.docker.com/engine/install/ubuntu/) site.

## Verify Docker
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

## Install Docker Compose
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

## Install Tools
Install the build, media streaming, and parsing tools:

   ```bash
   sudo apt install build-essential jq curl
   ```

NOTE: The device service ONLY works on Linux with kernel v5.10 or higher.  


## Tool Descriptions
The table below lists command line tools this guide uses to help with EdgeX configuration and device setup.

| Tool        | Description | Note |
| ----------- | ----------- |----------- |
| **build-essential** |  Developer tools such as libc, gcc, g++ and make. | |
| **curl**     | Allows the user to connect to services such as EdgeX. |Use curl to get transfer information either to or from this service. In the tutorial, use `curl` to communicate with the EdgeX API. The call will return a JSON object.|
| **jq**   |Parses the JSON object returned from the `curl` requests. |The `jq` command includes parameters that are used to parse and format data. In this tutorial, the `jq` command has been configured to return and format appropriate data for each `curl` command that is piped into it. |

>Table 1: Command Line Tools

###  Download EdgeX Compose Repository

1. Create a directory for the EdgeX compose repository:
   ```bash
   mkdir ~/edgex
   ```

2. Change into newly created directory:
   ```bash
   cd ~/edgex
   ```

3. Clone the EdgeX compose repository
   ```bash
   git clone https://github.com/edgexfoundry/edgex-compose.git
   ```


### Get the Device USB Camera Source Code

1. Change into the edgex directory:
   ```bash
   cd ~/edgex
   ```

2. Clone the device-usb-camera repository:

   ```bash
   git clone https://github.com/edgexfoundry/device-usb-camera.git
   ```


## Next Steps
[Build and run the software](./general-usage.md)  

## License
[Apache-2.0](LICENSE)
