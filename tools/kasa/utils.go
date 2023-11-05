package kasa

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

func transcode(in, out interface{}) {
	buf := new(bytes.Buffer)
	json.NewEncoder(buf).Encode(in)
	json.NewDecoder(buf).Decode(out)
}

func baseRequest(link TPLink, requestBody map[string]interface{}) (interface{}, error) {
	client := &http.Client{}
	params := &url.Values{
		"appName": {"Kasa_Android"},
		"termId":  {link.TermId()},
		"appVer":  {"1.4.4.607"},
		"ospf":    {"Android+6.0.1"},
		"netType": {"wifi"},
		"locale":  {"en_ES"},
	}
	if link.Token() != "" {
		params.Add("token", link.Token())
	}
	reqBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://wap.tplinkcloud.com/", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.URL.RawQuery = params.Encode()
	req.Header.Add("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 6.0.1; A0001 Build/M4B30X)")
	req.Header.Add("Content-Type", "application/json")
	response, err := client.Do(req)
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
		return nil, &LoginError{}
	}
	return res.Result, nil
}
