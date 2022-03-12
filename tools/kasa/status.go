package kasa

// For Device SysInfo
type ctrlProtocol struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type defaultOnState struct {
	Brightness int    `json:"brightness"`
	ColorTemp  int    `json:"color_temp"`
	Hue        int    `json:"hue"`
	Saturation int    `json:"saturation"`
	Mode       string `json:"mode"`
}

type PreferredState struct {
	defaultOnState
	Index int `json:"index"`
}

type sysInfoLightState struct {
	defaultOnState
	OnOff int `json:"on_off"`
}

type SysInfo struct {
	ActiveMode          string             `json:"active_mode"`
	Alias               string             `json:"alias"`
	CtrlProtocols       *ctrlProtocol      `json:"ctrl_protocols"`
	Description         string             `json:"description"`
	DeviceState         string             `json:"dev_state"`
	DeviceId            string             `json:"deviceId"`
	DiscoVersion        string             `json:"disco_ver"`
	ErrorCode           int                `json:"err_code"`
	HeapSize            int                `json:"heapsize"`
	HwId                string             `json:"hwId"`
	HwVer               string             `json:"hw_ver"`
	IsColor             int                `json:"is_color"`
	IsDimmable          int                `json:"is_dimmable"`
	IsFactory           bool               `json:"is_factory"`
	IsVariableColorTemp int                `json:"is_variable_color_temp"`
	LightState          *sysInfoLightState `json:"light_state"`
	MicMac              string             `json:"mic_mac"`
	MicType             string             `json:"mic_type"`
	Model               string             `json:"model"`
	OemId               string             `json:"oemId"`
	PreferredState      []*PreferredState  `json:"preferred_state"`
	RSSI                int                `json:"rssi"`
	SwVer               string             `json:"sw_ver"`
}

type sysInfoWrapper struct {
	SysInfo *SysInfo `json:"get_sysinfo"`
}

type sysInfoResponse struct {
	System *sysInfoWrapper `json:"system"`
}
