
<a name="USB Camera Device Service (found in device-usb-camera) Changelog"></a>
## Edgex USB Camera Device Service
[Github repository](https://github.com/edgexfoundry/device-usb-camera)

### Change Logs for EdgeX Dependencies

- [device-sdk-go](https://github.com/edgexfoundry/device-sdk-go/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md)  (indirect dependency)
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [v2.3.0] Levski - 2022-11-09 (Not Compatible with 1.x releases)

### Features ✨
- add internal command request/response topics ([#f7e3d81](https://github.com/edgexfoundry/device-usb-camera/commits/f7e3d81))
- **snap:** add config interface with unique identifier ([#103](https://github.com/edgexfoundry/device-usb-camera/issues/103)) ([#b346198](https://github.com/edgexfoundry/device-usb-camera/commits/b346198))

### Code Refactoring ♻
- **snap:** Simplify rtsp server versioning and configuration ([#1dbc6f5](https://github.com/edgexfoundry/device-usb-camera/commits/1dbc6f5))

### Bug Fixes 🐛
- improve error messaging on read and write commands ([#116](https://github.com/edgexfoundry/device-usb-camera/issues/116)) ([#a13c0d8](https://github.com/edgexfoundry/device-usb-camera/commits/a13c0d8))
- address pr comments and add timer ([#81c0ea8](https://github.com/edgexfoundry/device-usb-camera/commits/81c0ea8))
- return if stop error ([#679fd9a](https://github.com/edgexfoundry/device-usb-camera/commits/679fd9a))
- startStreaming now returns all errors ([#50aed43](https://github.com/edgexfoundry/device-usb-camera/commits/50aed43))
- update command request and response topic ([#345b3c5](https://github.com/edgexfoundry/device-usb-camera/commits/345b3c5))
- Correction in config field syntax ([#ae041b2](https://github.com/edgexfoundry/device-usb-camera/commits/ae041b2))

### Documentation 📖
- adding USB camera postman collection and env files ([#96](https://github.com/edgexfoundry/device-usb-camera/issues/96)) ([#e6cf2f2](https://github.com/edgexfoundry/device-usb-camera/commits/e6cf2f2))
- usb rtsp streaming guide and readme ([#1](https://github.com/edgexfoundry/device-usb-camera/issues/1)) ([#92](https://github.com/edgexfoundry/device-usb-camera/issues/92)) ([#2387317](https://github.com/edgexfoundry/device-usb-camera/commits/2387317))

### Build 👷
- Add option to build Service with NATS Capability ([#d8abada](https://github.com/edgexfoundry/device-usb-camera/commits/d8abada))
- Update to use latest SDK ([#d488376](https://github.com/edgexfoundry/device-usb-camera/commits/d488376))
- Latest go-mods, config and Dockerfile fix ([#59d67f4](https://github.com/edgexfoundry/device-usb-camera/commits/59d67f4))
- Upgrade to Go 1.18, Alpine 3.16, linter version and latest SDK/go-mod versions ([#0a9f00b](https://github.com/edgexfoundry/device-usb-camera/commits/0a9f00b))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#80752d2](https://github.com/edgexfoundry/device-usb-camera/commits/80752d2))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#8b7a325](https://github.com/edgexfoundry/device-usb-camera/commits/8b7a325))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#04b2efe](https://github.com/edgexfoundry/device-usb-camera/commits/04b2efe))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#c4bbce3](https://github.com/edgexfoundry/device-usb-camera/commits/c4bbce3))
- **deps:** Bump github.com/edgexfoundry/go-mod-core-contracts/v2 ([#06aed6e](https://github.com/edgexfoundry/device-usb-camera/commits/06aed6e))
- **deps:** Bump github.com/edgexfoundry/go-mod-core-contracts/v2 ([#37cb0c5](https://github.com/edgexfoundry/device-usb-camera/commits/37cb0c5))
- **deps:** Bump github.com/edgexfoundry/go-mod-core-contracts/v2 ([#106ba8d](https://github.com/edgexfoundry/device-usb-camera/commits/106ba8d))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#34076f9](https://github.com/edgexfoundry/device-usb-camera/commits/34076f9))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#5d7d954](https://github.com/edgexfoundry/device-usb-camera/commits/5d7d954))
- **deps:** Bump github.com/edgexfoundry/go-mod-core-contracts/v2 ([#dd7a03a](https://github.com/edgexfoundry/device-usb-camera/commits/dd7a03a))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#7e289ea](https://github.com/edgexfoundry/device-usb-camera/commits/7e289ea))
- **deps:** Bump github.com/edgexfoundry/device-sdk-go/v2 ([#3ae4a4f](https://github.com/edgexfoundry/device-usb-camera/commits/3ae4a4f))


## [v2.2.0] Kamakura - 2022-07-26

This is the initial release of this USB camera device service. Refer to the [README](https://github.com/edgexfoundry/device-usb-camera/blob/v2.2.0/README.md) for details about this service.