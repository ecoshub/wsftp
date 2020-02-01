package tools

import (
	"strings"
	"os/user"
	"net"
	"github.com/ecoshub/penman"
	"github.com/ecoshub/jint"
	"wsftp/log"
)
var (
	SEPARATOR  string
	MY_USERNAME  string
	MY_NICK  string
	MY_IP string
	MY_MAC  string
	BROADCAST_IP string
	SETTINGS_DIR  string
	UNKNOWN string
)

func init(){
	var err error
	UNKNOWN = "UNKNOWN"
	SEPARATOR = penman.Sep()
	SETTINGS_DIR = penman.GetHome() + SEPARATOR + "Documents" + SEPARATOR + "wsftp-settings.json"
	MY_USERNAME = GetUsername()
	MY_NICK = GetNick()
	MY_IP, err = GetInterfaceIP()
	if err != nil {
		// NEEDS MAJOR ERROR HANDLING
	}
	MY_MAC, err = GetMac()
	if err != nil {
		// NEEDS MAJOR ERROR HANDLING
	}
	BROADCAST_IP = GetBroadcastIP()
}

func GetUsername() string{
	user, err := user.Current()
	if err != nil {
		return UNKNOWN + "_USERNAME"
	}
	return user.Username
}

func GetNick() string{
	if penman.IsFileExist(SETTINGS_DIR) {
		if !penman.IsFileEmpty(SETTINGS_DIR) {
			file := penman.Read(SETTINGS_DIR)
			if username, done := jint.Get(file, "username") ; done == nil {
				return string(username)
			}
		}
	}
	return MY_USERNAME
}

func GetInterfaceIP() (string, error){
	Inters, _ := net.Interfaces()
	inslen := len(Inters)
	myAddr := ""
	for i := 0 ; i < inslen ; i++ {
		if Inters[i].Flags &  net.FlagLoopback != net.FlagLoopback && Inters[i].Flags & net.FlagUp == net.FlagUp{
			addr, _ := Inters[i].Addrs()
			if addr != nil {
				for _,ad := range addr{
					if strings.Contains(ad.String(), "."){
						myAddr = ad.String()
						break
					}
				}
				ip, _, _ := net.ParseCIDR(myAddr)
				return ip.String(), nil
			}
		}
	}
	return "0.0.0.0", log.INTERFACE_IP_RESOLVE_ERROR()
}

func GetMac() (string, error) {
	ins, _ := net.Interfaces()
	for _,in := range ins {
		if len(in.Name) > 0{
			if in.Name[0] == 'e' {
				return in.HardwareAddr.String(), nil
			}
		}
	}
	for _,in := range ins {
		if in.HardwareAddr.String() != ""{
			return in.HardwareAddr.String(), nil
		}
	}
	return UNKNOWN + "_MAC", log.INTERFACE_MAC_RESOLVE_ERROR()
}

func GetBroadcastIP() string{
	tokens := strings.Split(MY_IP, ".")
	tokens[len(tokens) - 1] = "255"
	return strings.Join(tokens, ".")
}