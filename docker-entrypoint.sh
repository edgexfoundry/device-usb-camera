#!/usr/bin/dumb-init /bin/sh

echo "Run rtsp-simple-server..."
/rtsp-simple-server &

echo "Run device-usb-camera..."
/device-usb-camera $@
