package log

import "errors"
import "fmt"

func WRONG_PORT_INDEX(at string) error {
	return errors.New(fmt.Sprintf("Fatal Error. Wrong port index. at:%v", at))
}
func MAIN_PORT_BUSSY() error {
	return errors.New("Fatal Error. The ports that required for the program to work properly is busy. Please close other program/programs that using this ports. Port range is [9997:10002]")
}
func INTERFACE_IP_RESOLVE_ERROR() error {
	return errors.New("Fatal Error. Interface IP resolve error. at: GetInterfaceIP()")
}
func INTERFACE_MAC_RESOLVE_ERROR() error {
	return errors.New("Fatal Error. Interface MAC resolve error. at: GetMac()")
}