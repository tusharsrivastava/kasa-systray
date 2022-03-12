package main

import (
	"github.com/tusharsrivastava/kasa-systray/tools"
)

func main() {
	var configuration = tools.SetupConfiguration()
	passphrase, err := tools.SetPassphraseGUI()
	if err != nil {
		panic(err)
	}
	_ = configuration.SetPassphrase(passphrase)
	app := tools.NewTray("", "Kasa by TPLink", configuration)
	app.Run()
}
