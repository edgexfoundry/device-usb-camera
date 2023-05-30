
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

## [3.0.0] Minnesota - 2023-05-31 (Only compatible with the 3.x releases)

### Features ✨
- Support for rtsp server authentication ([#240](https://github.com/edgexfoundry/device-usb-camera/issues/240)) ([#6884326](https://github.com/edgexfoundry/device-usb-camera/commits/6884326))
- **snap:** Copy provision watcher files in snapcraft ([#150](https://github.com/edgexfoundry/device-usb-camera/issues/150)) ([#d43df40](https://github.com/edgexfoundry/device-usb-camera/commits/d43df40))

### Bug Fixes 🐛
- Return ffmpeg error logs to caller, and fix StreamingStatus ([#254](https://github.com/edgexfoundry/device-usb-camera/issues/254)) ([#e4cb32a](https://github.com/edgexfoundry/device-usb-camera/commits/e4cb32a))
- Upgrade rtsp-simple-server to fix vulnerability ([#3d9796f](https://github.com/edgexfoundry/device-usb-camera/commits/3d9796f))
- **snap:** Refactor to avoid conflicts with readonly config provider directory ([#194](https://github.com/edgexfoundry/device-usb-camera/issues/194)) ([#8746c93](https://github.com/edgexfoundry/device-usb-camera/commits/8746c93))
- **snap:** Set snap-specific provision watchers directory ([#175](https://github.com/edgexfoundry/device-usb-camera/issues/175)) ([#4e47f77](https://github.com/edgexfoundry/device-usb-camera/commits/4e47f77))

### Code Refactoring ♻
- Consume Provision Watcher changes for running multiple instances ([#52b8227](https://github.com/edgexfoundry/device-usb-camera/commits/52b8227))
- Consume SDK interface changes
  ```text
  BREAKING CHANGE: Consume SDK interface changes by adding Start, Discover and ValidateDevice func on driver
  ```
- Change configuration and device toml files to yaml ([#a642c90](https://github.com/edgexfoundry/device-usb-camera/commits/a642c90))
  ```text
  BREAKING CHANGE: Configuration and device files now use yaml instead of toml
  ```
- Use device sdk for adding provision watchers and remove manual code ([#eb09eea](https://github.com/edgexfoundry/device-usb-camera/commits/eb09eea))
  ```text
  BREAKING CHANGE: Remove manual code to add provision watchers and instead use device-sdk to add them
  ```
- Replace internal topics from config with new constants ([#69957f4](https://github.com/edgexfoundry/device-usb-camera/commits/69957f4))
  ```text
  BREAKING CHANGE: Internal topics no longer configurable, except the base topic.
  ```
- Updates for common config
  ```text
  BREAKING CHANGE: Configuration file changed to remove common config settings
  ```
- Remove ZeroMQ message bus capability
  ```text
  BREAKING CHANGE: Remove ZeroMQ message bus capability
  ``` 
- Rework code for refactored MessageBus Configuration ([#bd8c447](https://github.com/edgexfoundry/device-usb-camera/commits/bd8c447))
   ```text
  BREAKING CHANGE: MessageQueue renamed to MessageBus and fields changed.
  ```
- Rename command line flags for the sake of consistency ([#11d8830](https://github.com/edgexfoundry/device-usb-camera/commits/11d8830))
  ```text
  BREAKING CHANGE: Renamed -c/--confdir to -cd/--configDir and -f/--file to -cf/--configFile
  ```
- Use latest SDK for flattened config stem ([#df2144b](https://github.com/edgexfoundry/device-usb-camera/commits/df2144b))
  ```text
  BREAKING CHANGE: Location of service configuration in Consul changed
  ```
- **snap:** Update command and metadata sourcing ([#190](https://github.com/edgexfoundry/device-usb-camera/issues/190)) ([#585c9f0](https://github.com/edgexfoundry/device-usb-camera/commits/585c9f0))
- **snap:** Refactor and upgrade to edgex-snap-hooks v3 ([#129](https://github.com/edgexfoundry/device-usb-camera/issues/129)) ([#ad81b67](https://github.com/edgexfoundry/device-usb-camera/commits/ad81b67))

### Documentation 📖
- Updated postman collection from v2 to v3 ([#251](https://github.com/edgexfoundry/device-usb-camera/issues/251)) ([#faedb97](https://github.com/edgexfoundry/device-usb-camera/commits/faedb97))
- Remove docs ([#238](https://github.com/edgexfoundry/device-usb-camera/issues/238)) ([#9fcc0da](https://github.com/edgexfoundry/device-usb-camera/commits/9fcc0da))
- Add warning to main branch and link to levski ([#191](https://github.com/edgexfoundry/device-usb-camera/issues/191)) ([#96327e0](https://github.com/edgexfoundry/device-usb-camera/commits/96327e0))
- Add warning for main branch and link to releases ([#185](https://github.com/edgexfoundry/device-usb-camera/issues/185)) ([#59664bc](https://github.com/edgexfoundry/device-usb-camera/commits/59664bc))
- Change location of nats documentation ([#168](https://github.com/edgexfoundry/device-usb-camera/issues/168)) ([#6531ea4](https://github.com/edgexfoundry/device-usb-camera/commits/6531ea4))
- **snap:** Update camera interface, getting started doc version ([#143](https://github.com/edgexfoundry/device-usb-camera/issues/143)) ([#c4833ec](https://github.com/edgexfoundry/device-usb-camera/commits/c4833ec))

### Build 👷
- Ignore all go-mods except device-sdk-go ([#8f2ed8f](https://github.com/edgexfoundry/device-usb-camera/commits/8f2ed8f))
- Update to use latest sdk ([#174](https://github.com/edgexfoundry/device-usb-camera/issues/174)) ([#3323a63](https://github.com/edgexfoundry/device-usb-camera/commits/3323a63))
- Update to Go 1.20, Alpine 3.17 and linter v1.51.2 ([#90e2dc0](https://github.com/edgexfoundry/device-usb-camera/commits/90e2dc0))
- Ignore all go-mods except device-sdk-go ([#1a7f20c](https://github.com/edgexfoundry/device-usb-camera/commits/1a7f20c))


## [v2.3.0] Levski - 2022-11-09 (Not Compatible with 1.x releases)

### Features ✨
- implement rediscovery ([#120](https://github.com/edgexfoundry/device-usb-camera/pull/120)) ([#9bcd451](https://github.com/edgexfoundry/device-usb-camera/commit/9bcd451))
- add internal command request/response topics ([#f7e3d81](https://github.com/edgexfoundry/device-usb-camera/commits/f7e3d81))
- **snap:** add config interface with unique identifier ([#103](https://github.com/edgexfoundry/device-usb-camera/issues/103)) ([#b346198](https://github.com/edgexfoundry/device-usb-camera/commits/b346198))

### Code Refactoring ♻
- **snap:** Simplify rtsp server versioning and configuration ([#1dbc6f5](https://github.com/edgexfoundry/device-usb-camera/commits/1dbc6f5))

### Bug Fixes 🐛
- improve error messaging for incorrect protocol properties ([#117](https://github.com/edgexfoundry/device-usb-camera/issues/117)) ([#7dbe31c](https://github.com/edgexfoundry/device-usb-camera/commits/7dbe31c))  
- improve error messaging on read and write commands ([#116](https://github.com/edgexfoundry/device-usb-camera/issues/116)) ([#a13c0d8](https://github.com/edgexfoundry/device-usb-camera/commits/a13c0d8))
- error forwarding in startStreaming ([#113](https://github.com/edgexfoundry/device-usb-camera/issues/113)) ([#02bc335](https://github.com/edgexfoundry/device-usb-camera/commit/02bc3351eb583ffe88737b5638435757cc287900)) ([#81c0ea8](https://github.com/edgexfoundry/device-usb-camera/commits/81c0ea8)) ([#679fd9a](https://github.com/edgexfoundry/device-usb-camera/commits/679fd9a)) ([#50aed43](https://github.com/edgexfoundry/device-usb-camera/commits/50aed43fc5ea9f2235be704591a04f41aa30b17f))
- update command request and response topic ([#345b3c5](https://github.com/edgexfoundry/device-usb-camera/commits/345b3c5)) 
- Correction in config field syntax ([#ae041b2](https://github.com/edgexfoundry/device-usb-camera/commits/ae041b2))

### Documentation 📖
- updated change log for levski release ([#118](https://github.com/edgexfoundry/device-usb-camera/issues/118)) ([#e17bf6e](https://github.com/edgexfoundry/device-usb-camera/commits/e17bf6e))
- updates to usb documentation  ([#115](https://github.com/edgexfoundry/device-usb-camera/issues/115)) ([#d57e067](https://github.com/edgexfoundry/device-usb-camera/commits/d57e067))
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
