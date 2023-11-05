package kasa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type TPLinkDeviceInfo struct {
	FwVer        string `json:"fwVer"`
	Alias        string `json:"alias"`
	Status       int    `json:"status"`
	Role         string `json:"role"`
	DeviceId     string `json:"deviceId"`
	DeviceMac    string `json:"deviceMac"`
	DeviceName   string `json:"deviceName"`
	DeviceType   string `json:"deviceType"`
	DeviceModel  string `json:"deviceModel"`
	AppServerUrl string `json:"appServerUrl"`
}

type Device interface {
	Id() string
	FirmwareVersion() string
	Role() string
	Mac() string
	Model() string
	Name() string
	Type() string
	Status() int
	Alias() string
	AppServerUrl() string
	HumanName() string
	Brightness() int
	IsConnected() bool
	IsDisconnected() bool
	TurnOn() error
	TurnOff() error
	SystemInfo() (*SysInfo, error)
	PreferredStates() []*PreferredState
	SetPreferredState(idx int) error
	passthroughRequest(command map[string]interface{}) (map[string]interface{}, error)
}

type TpLinkDevice struct {
	GenericType     string
	device          *TPLinkDeviceInfo
	params          *url.Values
	preferredStates []*PreferredState
	brightness      int
}

func NewTpLinkDevice(link TPLink, deviceInfo *TPLinkDeviceInfo) Device {
	dev := &TpLinkDevice{
		GenericType: "device",
		device:      deviceInfo,
		brightness:  0,
		params: &url.Values{
			"appName": {"Kasa_Android"},
			"termID":  {link.TermId()},
			"appVer":  {"1.4.4.607"},
			"ospf":    {"Android+6.0.1"},
			"netType": {"wifi"},
			"locale":  {"es_ES"},
			"token":   {link.Token()},
		},
	}
	dev.syncState()
	return dev
}

func (d *TpLinkDevice) PreferredStates() []*PreferredState {
	return d.preferredStates
}

func (d *TpLinkDevice) Id() string {
	return d.device.DeviceId
}

func (d *TpLinkDevice) FirmwareVersion() string {
	return d.device.FwVer
}

func (d *TpLinkDevice) Role() string {
	return d.device.Role
}

func (d *TpLinkDevice) Mac() string {
	return d.device.DeviceMac
}

func (d *TpLinkDevice) Model() string {
	return d.device.DeviceModel
}

func (d *TpLinkDevice) Name() string {
	return d.device.DeviceName
}

func (d *TpLinkDevice) Type() string {
	return d.device.DeviceType
}

func (d *TpLinkDevice) Status() int {
	return d.device.Status
}

func (d *TpLinkDevice) Alias() string {
	return d.device.Alias
}

func (d *TpLinkDevice) AppServerUrl() string {
	return d.device.AppServerUrl
}

func (d *TpLinkDevice) HumanName() string {
	status := func(status bool) string {
		if status {
			return "ON"
		}
		return "OFF"
	}(d.IsConnected())

	return fmt.Sprintf("%s [%s %d%%]", d.device.Alias, status, d.brightness)
}

func (d *TpLinkDevice) Brightness() int {
	return d.brightness
}

func (d *TpLinkDevice) IsConnected() bool {
	return d.device.Status == 1
}

func (d *TpLinkDevice) IsDisconnected() bool {
	return d.device.Status == 0
}

func (d *TpLinkDevice) TurnOn() error {
	_, err := d.passthroughRequest(map[string]interface{}{
		"smartlife.iot.smartbulb.lightingservice": map[string]interface{}{
			"transition_light_state": map[string]interface{}{
				"brightness": 100,
				"on_off":     1,
			},
		},
	})
	if err != nil {
		return err
	}
	return d.syncState()
}

func (d *TpLinkDevice) TurnOff() error {
	_, err := d.passthroughRequest(map[string]interface{}{
		"smartlife.iot.smartbulb.lightingservice": map[string]interface{}{
			"transition_light_state": map[string]interface{}{
				"brightness": 100,
				"on_off":     0,
			},
		},
	})
	if err != nil {
		return err
	}
	return d.syncState()
}

func (d *TpLinkDevice) SystemInfo() (*SysInfo, error) {
	sysInfo, err := d.passthroughRequest(map[string]interface{}{
		"system": map[string]interface{}{
			"get_sysinfo": map[string]interface{}{},
		},
	})
	if err != nil {
		return nil, err
	}
	response := &sysInfoResponse{}
	transcode(sysInfo, &response)
	return response.System.SysInfo, nil
}

func (d *TpLinkDevice) SetPreferredState(idx int) error {
	if idx < 0 || idx >= len(d.preferredStates) {
		return fmt.Errorf("invalid preferred state index %d", idx)
	}
	state := d.preferredStates[idx]
	_, err := d.passthroughRequest(map[string]interface{}{
		"smartlife.iot.smartbulb.lightingservice": map[string]interface{}{
			"transition_light_state": map[string]interface{}{
				"brightness": state.Brightness,
				"on_off":     1,
			},
		},
	})
	if err != nil {
		return err
	}
	return d.syncState()
}

func (d *TpLinkDevice) passthroughRequest(command map[string]interface{}) (map[string]interface{}, error) {
	cmdJson, _ := json.Marshal(command)
	requestBody, _ := json.Marshal(map[string]interface{}{
		"method": "passthrough",
		"params": map[string]string{
			"deviceId":    d.Id(),
			"requestData": string(cmdJson),
		},
	})
	client := &http.Client{}
	request, err := http.NewRequest("POST", d.AppServerUrl(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	request.Header = http.Header{
		"cache-control": []string{"no-cache"},
		"User-Agent":    []string{"Dalvik/2.1.0 (Linux; U; Android 6.0.1; A0001 Build/M4B30X)"},
		"Content-Type":  []string{"application/json"},
	}
	request.URL.RawQuery = d.params.Encode()
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	res := &Response{}
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, err
	}

	if res.ErrorCode != 0 {
		return nil, errors.New(res.Message)
	}

	responseData := res.Result.(map[string]interface{})["responseData"]
	if responseData != "" {
		var data map[string]interface{}
		err = json.Unmarshal([]byte(responseData.(string)), &data)
		if err != nil {
			return nil, err
		}
		return data, nil
	}
	return res.Result.(map[string]interface{}), nil
}

func (d *TpLinkDevice) syncState() error {
	sysInfo, err := d.SystemInfo()
	if err != nil {
		return err
	}
	devInfo := &TPLinkDeviceInfo{
		FwVer:        d.device.FwVer,
		Alias:        sysInfo.Alias,
		Status:       sysInfo.LightState.OnOff,
		Role:         d.device.Role,
		DeviceId:     sysInfo.DeviceId,
		DeviceMac:    sysInfo.MicMac,
		DeviceName:   sysInfo.Description,
		DeviceType:   sysInfo.MicType,
		DeviceModel:  sysInfo.Model,
		AppServerUrl: d.device.AppServerUrl,
	}
	d.device = devInfo
	d.brightness = sysInfo.LightState.Brightness
	d.preferredStates = sysInfo.PreferredState
	return nil
}
