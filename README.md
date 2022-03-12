# TPLink Kasa Systray Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/tusharsrivastava/kasa-systray)](https://goreportcard.com/report/github.com/tusharsrivastava/kasa-systray)

A simple systray application for TPLink Kasa Smart devices. Currently it only supports the Smart Bulb but the plan is to add support for other devices as well. This application could not have been possible without the help of the following libraries:

- [tplink-cloud-api](https://github.com/adumont/tplink-cloud-api): This library written in node.js was used to write the kasa API layer. _Note: A lot of work is required as currently I only wrote the API layer for the Smart Bulb._
- [getlantern/systray](https://github.com/getlantern/systray): This library is the heart of systray and thanks to author, is cross platform.
- [zalando/go-keyring](https://github.com/zalando/go-keyring): This library provides the keyring functionality for storing the passphrase used to encrypt and store credentials in config.json file.
- [ncruces/zenity](https://github.com/ncruces/zenity): This library provides the GUI layer for the application and is cross platform. It is used to display notifications, prompts for user inputs and to display error dialogs.
- [spf13/viper](https://github.com/spf13/viper): This library provides the configuration file parsing and validation.