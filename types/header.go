package types

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

type YiduiHeaders http.Header

func (head YiduiHeaders) OsType() string {
	chName := head.ChannelName()
	switch chName {
	case "market_appstore":
		return "ios"
	case "mini_app":
		return chName
	default:
		if ok, _ := regexp.MatchString("(Yidui-Android)+", head.ClientUa()); ok {
			return "android"
		}
		return ""
	}
}

func (head YiduiHeaders) GetApiKey() string {
	if key := head.Get("Apikey"); key != "" {
		return key
	}
	return "7e08df24"
}

func (head YiduiHeaders) CodeTag() string {
	return strings.ReplaceAll(head.Get("Codetag"), "_", "-")
}

func (head YiduiHeaders) ClientAndroidID() string {
	return head.Get("Android-Id")
}
func (head YiduiHeaders) GetAndroidId() string {
	rcSign := head.Get("Rcsign")
	if rcSign == "" {
		rcSign = head.Get("RcSign")
	}
	if rcSign == "" {
		return ""
	}

	temp := map[string]string{}
	if err := json.Unmarshal([]byte(rcSign), &temp); err == nil {
		if v, ok := temp["android_id"]; ok {
			return v
		}
	}
	return ""
}
func (head YiduiHeaders) OsVersion() string {
	return head.Get("Osversion")
}

func (head YiduiHeaders) ScDeviceID() string {
	if head.ChannelName() == "market_appstore" {
		return head.DeviceID()
	}
	return head.ClientAndroidID()
}

func (head YiduiHeaders) YiDuiDeviceID() string {
	return head.Get("Deviceid")
}

func (head YiduiHeaders) GioID() string {
	return head.Get("Giouid")
}

func (head YiduiHeaders) DeviceID() string {
	if m := head.Get("Imei"); m != "" {
		return m
	}
	return head.YiDuiDeviceID()
}

func (head YiduiHeaders) IMei() string {
	v := head.Get("Imei")
	if v == "" {
		v = head.Get("IMEI")
	}
	return v
}

func (head YiduiHeaders) Authorization() string {
	return head.Get("Authorization")
}

func (head YiduiHeaders) ChannelName() string {
	for _, cName := range []string{"Channel", "channel"} {
		if ch := head.Get(cName); ch != "" {
			return ch
		}
	}
	return ""
}

func (head YiduiHeaders) ClientUa() string {
	return head.Get("User-Agent")
}

func (head YiduiHeaders) Uri() string {
	return head.Get("Uri")
}

func (head YiduiHeaders) Brand() string {
	return head.Get("Brand")
}

func (head YiduiHeaders) ClientOaID() string {
	return head.Get("Oaid")
}

func (head YiduiHeaders) ClientIP() string {
	if str, ok := head["X-Forwarded-For"]; ok && len(str) > 0 {
		values := strings.Split(str[0], ",")
		if len(values) > 0 {
			return values[0]
		}
		return head.Get("request_ip")
	}
	return head.Get("request_ip")
}

func (head YiduiHeaders) Get(key string) string {
	if str, ok := head[key]; ok && len(str) > 0 {
		return str[0]
	}
	return ""
}

func (head YiduiHeaders) Set(key, value string) {
	head[key] = []string{value}
}
