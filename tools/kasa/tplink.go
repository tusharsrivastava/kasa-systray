package kasa

import (
	"log"

	"github.com/google/uuid"
)

type TPLink interface {
	TermId() string
	Token() string
	DeviceList() []Device
	FindDevice(alias string) (Device, error)
}

type tpLink struct {
	termId  string
	token   string
	devices []Device
}

func (t *tpLink) TermId() string {
	return t.termId
}

func (t *tpLink) Token() string {
	return t.token
}

func (t *tpLink) DeviceList() []Device {
	command := map[string]interface{}{"method": "getDeviceList"}
	res, err := baseRequest(t, command)
	if err != nil {
		log.Println(err)
		return nil
	}
	for i, v := range res.(map[string]interface{}) {
		if i == "deviceList" {
			for _, v := range v.([]interface{}) {
				var device TPLinkDeviceInfo
				transcode(v, &device)
				dev := NewTpLinkDevice(t, &device)
				t.devices = append(t.devices, dev)
			}
		}
	}
	return t.devices
}

func (t *tpLink) FindDevice(alias string) (Device, error) {
	for _, device := range t.devices {
		if device.Alias() == alias {
			return device, nil
		}
	}
	return nil, nil
}

func TpLinkLogin(username string, password string) (TPLink, error) {
	termId := uuid.New()
	link := &tpLink{
		termId:  termId.String(),
		token:   "",
		devices: nil,
	}
	reqBody := map[string]interface{}{
		"method": "login",
		"url":    "https://wap.tplinkcloud.com",
		"params": map[string]string{
			"appType":       "Kasa_Android",
			"cloudUserName": username,
			"cloudPassword": password,
			"terminalUUID":  termId.String(),
		},
	}

	res, err := baseRequest(link, reqBody)
	if err != nil {
		return nil, err
	}
	var loginResp LoginResponse
	transcode(res, &loginResp)
	return &tpLink{
		termId:  termId.String(),
		token:   loginResp.Token,
		devices: nil,
	}, nil
}
