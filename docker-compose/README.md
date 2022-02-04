The files in this directory are temporary and will be used to add device-usb-camera into [edgex-compose](https://github.com/edgexfoundry/edgex-compose) in the near future.

> **Note to Developers**: 
> *If you want to test the dockerized device-usb-camera in this stage, please download [edgex-compose](https://github.com/edgexfoundry/edgex-compose) and follow the instructions below to generate a compose file including EdgeX Core services and device-usb-camera service.*

1. Copy `add-device-usb-camera.yml` from this directory to edgex-compose/compose-builder.

2. Go to edgex-compose/compose-builder.

3. Update `.env` file to add the registry and image version variable for device-usb-camera.
```
DEVICE_USBCAM_VERSION=latest
```
4. Update Makefile to add new option `ds-usbcam` to the [OPTIONS list](https://github.com/edgexfoundry/edgex-compose/blob/main/compose-builder/Makefile#L38).

5. Update Makefile to add new device device-usb-camera. Search for keyword `# Add Device Services` and then add the following code snippet below that line.
```
ifeq (ds-usbcam, $(filter ds-usbcam,$(ARGS)))
	COMPOSE_FILES:=$(COMPOSE_FILES) -f add-device-usb-camera.yml
	ifeq (mqtt-bus, $(filter mqtt-bus,$(ARGS)))
	  extension_file:= $(shell GEN_EXT_DIR="$(GEN_EXT_DIR)" ./gen_mqtt_messagebus_compose_ext.sh device-usb-camera -d)
	  COMPOSE_FILES:=$(COMPOSE_FILES) -f $(extension_file)
	endif
	ifneq (no-secty, $(filter no-secty,$(ARGS)))
		ifeq ($(TOKEN_LIST),"")
			TOKEN_LIST:=device-usb-camera
		else
			TOKEN_LIST:=$(TOKEN_LIST),device-usb-camera
		endif
		ifeq ($(KNOWN_SECRETS_LIST),"")
			KNOWN_SECRETS_LIST:=redisdb[device-usb-camera]
		else
			KNOWN_SECRETS_LIST:=$(KNOWN_SECRETS_LIST),redisdb[device-usb-camera]
		endif
		extension_file:= $(shell GEN_EXT_DIR="$(GEN_EXT_DIR)" ./gen_secure_compose_ext.sh device-usb-camera)
		COMPOSE_FILES:=$(COMPOSE_FILES) -f $(extension_file)
	endif
endif
```

6. Open a terminal and go to edgex-compose/compose-builder, enter the command `make gen no-secty ds-usbcam` to generate non-secure compose or `make gen ds-usbcam` to generate secure compose.

7. Update the compose file just created (`docker-compose.yml`) to specify the locally-built image of device-usb-camera.
