// Tools Tray for Systray
package tools

import (
	"fmt"
	"log"
	"strconv"

	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"github.com/tusharsrivastava/kasa-systray/icon"
	"github.com/tusharsrivastava/kasa-systray/tools/kasa"
)

type Tray interface {
	Run()
	ready()
}

type devSubMenu struct {
	id   string
	menu *systray.MenuItem
}

type deviceMenu struct {
	device  kasa.Device
	menu    *systray.MenuItem
	submenu []*devSubMenu
}

type tray struct {
	title       string
	tooltip     string
	config      *Configuration
	devHolder   *systray.MenuItem
	devicesMenu map[string]*deviceMenu
}

func (t *tray) Run() {
	systray.Run(t.ready, nil)
}

func (t *tray) ready() {
	systray.SetIcon(icon.Data)
	systray.SetTitle(t.title)
	systray.SetTooltip(t.tooltip)
	go t.loop()
}

func (t *tray) loop() {
	login := systray.AddMenuItem("Login", "Login to TPLink")
	t.devHolder = systray.AddMenuItem("Devices", "Devices")
	t.devHolder.Disable()
	autoConnect := systray.AddMenuItemCheckbox(t.getAutoConnectTitle(), "Auto Connect", t.config.AutoConnect)
	mReset := systray.AddMenuItem("Reset", "Reset")
	mQuit := systray.AddMenuItem("Quit", "Quit")
	loginEvt := t.loginEvent(login)

	// Run the event handler goroutines
	go t.quitHandler(mQuit.ClickedCh)
	go t.resetHandler(mReset, mQuit.ClickedCh)
	go t.loginHandler(login, loginEvt)
	go t.autoConnectHandler(autoConnect, loginEvt)
}

func (t *tray) getAutoConnectTitle() string {
	if t.config.AutoConnect {
		return "Turn off Auto Connect"
	}
	return "Turn on Auto Connect"
}

func (t *tray) loginEvent(loginMenu *systray.MenuItem) chan bool {
	ch := make(chan bool)
	go func() {
		if t.config.AutoConnect {
			ch <- true
		}
		for {
			<-loginMenu.ClickedCh
			ch <- true
		}
	}()
	return ch
}

func (t *tray) quitHandler(ch chan struct{}) {
	<-ch
	systray.Quit()
}

func (t *tray) resetHandler(mReset *systray.MenuItem, quitChan chan struct{}) {
	for {
		<-mReset.ClickedCh
		err := ResetAll(t.config)
		if err != nil {
			DisplayErrorGUI(err)
		} else {
			quitChan <- struct{}{}
		}
	}
}

func (t *tray) loginHandler(login *systray.MenuItem, eventChan chan bool) {
	for {
		<-eventChan
		auth, isFresh, err := t.config.ReadAuth(true)
		if err != nil {
			DisplayErrorGUI(err)
			continue
		}
		link, err := kasa.TpLinkLogin(auth.Username, auth.Password)
		if err != nil {
			DisplayErrorGUI(err)
			continue
		}
		if isFresh {
			t.config.WriteConfiguration()
		}
		login.Disable()
		Notify("Logged in", "You are now logged in", zenity.InfoIcon)
		t.devHolder.Enable()
		devices := link.DeviceList()
		msg := fmt.Sprintf("Found %d device(s)", len(devices))
		Notify("Kasa Notify", msg, zenity.InfoIcon)
		t.createDevicesMenu(devices)
	}
}

func (t *tray) autoConnectHandler(autoConnect *systray.MenuItem, loginEventChan chan bool) {
	for {
		<-autoConnect.ClickedCh
		notifyMsg := "Auto Connect is now "
		if autoConnect.Checked() {
			autoConnect.Uncheck()
			notifyMsg += "off"
		} else {
			autoConnect.Check()
			notifyMsg += "on"
		}
		t.config.AutoConnect = autoConnect.Checked()
		err := t.config.WriteConfiguration()
		if err != nil {
			Notify("Kasa Error", err.Error(), zenity.ErrorIcon)
			log.Println(err)
		}
		autoConnect.SetTitle(t.getAutoConnectTitle())
		Notify("Kasa Notify", notifyMsg, zenity.InfoIcon)
		if autoConnect.Checked() {
			loginEventChan <- true
		}
	}
}

func (t *tray) createDevicesMenu(devices []kasa.Device) {
	for _, device := range devices {
		log.Println(device.HumanName())
		mainMenu := t.devHolder.AddSubMenuItem(device.HumanName(), device.Name())
		submenu := []*devSubMenu{}
		turnOn := mainMenu.AddSubMenuItem("Turn On", "Turn On")
		submenu = append(submenu, &devSubMenu{"on", turnOn})
		turnOff := mainMenu.AddSubMenuItem("Turn Off", "Turn Off")
		submenu = append(submenu, &devSubMenu{"off", turnOff})
		// Build Preferred State submenu
		for _, state := range device.PreferredStates() {
			title := fmt.Sprintf("Brightness [%d%%]", state.Brightness)
			prefState := mainMenu.AddSubMenuItem(title, fmt.Sprint(state.Brightness))
			submenu = append(submenu, &devSubMenu{fmt.Sprint(state.Index), prefState})
		}
		devMenu := &deviceMenu{device, mainMenu, submenu}
		t.devicesMenu[device.Id()] = devMenu

		if device.IsConnected() {
			turnOn.Disable()
			turnOff.Enable()
		} else {
			turnOn.Enable()
			turnOff.Disable()
		}
	}
	for _, dMenu := range t.devicesMenu {
		go func(dMenu *deviceMenu) {
			localCh := getSubmenuClickEvent(dMenu.submenu)
			for {
				sm := <-localCh
				switch sm.id {
				case "on":
					log.Println("Turning on")
					err := dMenu.device.TurnOn()
					if err != nil {
						Notify("Kasa Error", err.Error(), zenity.ErrorIcon)
						continue
					}
					msg := fmt.Sprintf("%s now turned On with %d%% brightness", dMenu.device.Alias(), dMenu.device.Brightness())
					Notify("Kasa Notify", msg, zenity.InfoIcon)
					log.Println("Turned on")
				case "off":
					log.Println("Turning off")
					err := dMenu.device.TurnOff()
					if err != nil {
						Notify("Kasa Error", err.Error(), zenity.ErrorIcon)
						continue
					}
					msg := fmt.Sprintf("%s now turned Off", dMenu.device.Alias())
					Notify("Kasa Notify", msg, zenity.InfoIcon)
					log.Println("Turned off")
				default:
					// Set preferred state
					log.Println("Setting preferred state")
					idx, err := strconv.ParseInt(sm.id, 10, 64)
					if err != nil {
						Notify("Kasa Error", err.Error(), zenity.ErrorIcon)
						continue
					}
					err = dMenu.device.SetPreferredState(int(idx))
					if err != nil {
						Notify("Kasa Error", err.Error(), zenity.ErrorIcon)
						continue
					}
					msg := fmt.Sprintf("%s now set to brightness %d%%", dMenu.device.Alias(), dMenu.device.Brightness())
					Notify("Kasa Notify", msg, zenity.InfoIcon)
					log.Println("Set preferred state")
				}
				for _, s := range dMenu.submenu {
					if s.id == "on" && dMenu.device.IsConnected() {
						s.menu.Disable()
					} else if s.id == "off" && dMenu.device.IsDisconnected() {
						s.menu.Disable()
					} else {
						s.menu.Enable()
					}
				}
				dMenu.menu.SetTitle(dMenu.device.HumanName())
			}
		}(dMenu)
	}
}

func getSubmenuClickEvent(menu []*devSubMenu) chan *devSubMenu {
	ch := make(chan *devSubMenu)
	for _, sm := range menu {
		go func(sm *devSubMenu) {
			for {
				<-sm.menu.ClickedCh
				ch <- sm
			}
		}(sm)
	}
	return ch
}

func NewTray(title string, tooltip string, config *Configuration) Tray {
	return &tray{title, tooltip, config, nil, map[string]*deviceMenu{}}
}
