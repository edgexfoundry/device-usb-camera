#
# Copyright (c) 2023 Intel Corporation
# Copyright (c) 2024 IOTech Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG BASE=golang:1.23-alpine3.20
FROM ${BASE} AS builder

ARG ADD_BUILD_TAGS=""
ARG MAKE="make -e ADD_BUILD_TAGS=$ADD_BUILD_TAGS build"

ARG ALPINE_PKG_BASE="make git gcc libc-dev curl"
ARG ALPINE_PKG_EXTRA="v4l-utils-dev v4l-utils v4l-utils-libs linux-headers"

RUN apk add --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}

WORKDIR /device-usb-camera

COPY go.mod vendor* ./
RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

COPY . .

RUN curl -o LICENSE-mediamtx https://raw.githubusercontent.com/bluenviron/mediamtx/main/LICENSE

RUN ${MAKE}

FROM bluenviron/mediamtx:1.15.3 AS rtsp

FROM alpine:3.20

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2023: Intel Corporation'

# dumb-init needed for injected secure bootstrapping entrypoint script when run in secure mode.
RUN apk add --update --no-cache dumb-init ffmpeg udev
# Ensure using latest versions of all installed packages to avoid any recent CVEs
RUN apk --no-cache upgrade

WORKDIR /
COPY --from=builder /device-usb-camera/cmd /
COPY --from=builder /device-usb-camera/LICENSE /
COPY --from=builder /device-usb-camera/LICENSE-mediamtx /
COPY --from=builder /device-usb-camera/Attribution.txt /
COPY --from=builder /device-usb-camera/docker-entrypoint.sh /
COPY --from=rtsp /mediamtx.yml /
COPY --from=rtsp /mediamtx /

# disable unused rtsp-server listeners
RUN sed -i 's/rtmp: no/rtmp: yes/g' mediamtx.yml
RUN sed -i 's/hls: no/hls: yes/g' mediamtx.yml
RUN sed -i 's/protocols: \[udp, multicast, tcp\]/protocols: \[tcp\]/g' mediamtx.yml
RUN sed -i 's,authHTTPAddress:,authHTTPAddress: http://localhost:8000/rtspauth,g' mediamtx.yml

EXPOSE 59983
# RTSP port of the internal rtsp-server:
EXPOSE 8554

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["-cp=keeper.http://edgex-core-keeper:59890", "--registry" ]
