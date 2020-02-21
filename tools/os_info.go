package tools

import (
	"errors"
	"github.com/ecoshub/jin"
	"github.com/ecoshub/penman"
	"net"
	"os/user"
	"strings"
)

var (
	UNKNOWN      string = "UNKNOWN"
	SEPARATOR    string = penman.Sep()
	SETTINGS_DIR string = penman.GetHome() + SEPARATOR + "Documents" + SEPARATOR + "wsftp-settings.json"
	MY_USERNAME  string = GetUsername()
	MY_NICK      string = GetNick()
	MY_IP        string
	MY_MAC       string
	BROADCAST_IP string

	ERROR_GET_IP  string = "Get: IP resolve error"
	ERROR_GET_MAC string = "Get: MAC resolve error"
	ERROR_NO_IP   string = "Fatal Error: No interface IP has found"
	ERROR_NO_MAC  string = "Fatal Error: No interface MAC has found"
)

func init() {
	var err error
	MY_IP, err = GetInterfaceIP()
	if err != nil {
		StdoutHandle("error", ERROR_GET_IP, err)
	}
	MY_MAC, err = GetMac()
	if err != nil {
		StdoutHandle("error", ERROR_GET_MAC, err)
	}
	BROADCAST_IP = GetBroadcastIP()
}

func GetUsername() string {
	user, err := user.Current()
	if err != nil {
		return UNKNOWN + "_USERNAME"
	}
	return user.Username
}

func GetNick() string {
	if penman.IsFileExist(SETTINGS_DIR) {
		if !penman.IsFileEmpty(SETTINGS_DIR) {
			file := penman.Read(SETTINGS_DIR)
			if username, done := jin.Get(file, "username"); done == nil {
				return string(username)
			}
		}
	}
	return MY_USERNAME
}

func GetInterfaceIP() (string, error) {
	Inters, _ := net.Interfaces()
	inslen := len(Inters)
	myAddr := ""
	for i := 0; i < inslen; i++ {
		if Inters[i].Flags&net.FlagLoopback != net.FlagLoopback && Inters[i].Flags&net.FlagUp == net.FlagUp {
			addr, _ := Inters[i].Addrs()
			if addr != nil {
				for _, ad := range addr {
					if strings.Contains(ad.String(), ".") {
						myAddr = ad.String()
						break
					}
				}
				ip, _, _ := net.ParseCIDR(myAddr)
				return ip.String(), nil
			}
		}
	}
	return "0.0.0.0", errors.New(ERROR_NO_IP)
}

func GetMac() (string, error) {
	ins, _ := net.Interfaces()
	for _, in := range ins {
		if len(in.Name) > 0 {
			if in.Name[0] == 'e' {
				return in.HardwareAddr.String(), nil
			}
		}
	}
	for _, in := range ins {
		if in.HardwareAddr.String() != "" {
			return in.HardwareAddr.String(), nil
		}
	}
	return UNKNOWN + "_MAC", errors.New(ERROR_NO_MAC)
}

func GetBroadcastIP() string {
	tokens := strings.Split(MY_IP, ".")
	tokens[len(tokens)-1] = "255"
	return strings.Join(tokens, ".")
}
