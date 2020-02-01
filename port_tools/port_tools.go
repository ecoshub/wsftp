package port_tools

import "strconv"
import "net"
import "wsftp2/locals"
import "wsftp2/log"

var (
	activeTransaction int    = 0
	ports                    = make([][]int, locals.ACTIVE_TRANSACTION_LIMIT)
	portIDMap                = make(map[int]string, locals.ACTIVE_TRANSACTION_LIMIT)
)
func init(){
	for i := 0; i < locals.ACTIVE_TRANSACTION_LIMIT; i++ {
		if PortCheck(locals.TCP_TRANSECTION_START_PORT - i) {
			ports[i] = []int{locals.TCP_TRANSECTION_START_PORT - i, 0}
		}
	}
}

func AllocatePort() int {
	for i := 0; i < locals.ACTIVE_TRANSACTION_LIMIT; i++ {
		if ports[i][1] == 0 && PortCheck(ports[i][0]) {
			ports[i][1] = 1
			return ports[i][0]
		}
	}
	return -1
}

func MainPortCheck() error{
	result := PortCheck(locals.WS_COMMANDER_LISTEN_PORT)
	if !result {return log.MAIN_PORT_BUSSY()}
	result = result && PortCheck(locals.UDP_HANDSHAKE_LISTEN_PORT)
	if !result {return log.MAIN_PORT_BUSSY()}
	result = result && PortCheck(locals.WS_HANDSHAKE_LISTEN_PORT)
	if !result {return log.MAIN_PORT_BUSSY()}
	result = result && PortCheck(locals.WS_SEND_RECEIVE_LISTEN_PORT)
	if !result {return log.MAIN_PORT_BUSSY()}
	result = result && PortCheck(locals.WS_MESSAGE_LISTEN_PORT)
	if !result {return log.MAIN_PORT_BUSSY()}
	return nil
}

func PortCheck(port int) bool {
	listener, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	defer listener.Close()
	if err != nil {
		return false
	}
	return true
}

func GetPortIndex(port int) (int, error) {
	for i := 0; i < locals.ACTIVE_TRANSACTION_LIMIT; i++ {
		if ports[i][0] == port {
			return i, nil
		}
	}
	return -1, log.WRONG_PORT_INDEX("GetPortIndex()")
}

func SetPortBusy(port int) (bool, error) {
	index, err := GetPortIndex(port)
	if err != nil {
		return false, err
	}
	if index > -1 && index < locals.ACTIVE_TRANSACTION_LIMIT {
		ports[index][1] = 1
		return true, nil
	}
	return false, log.WRONG_PORT_INDEX("SetPortBusy()")
}

func FreePort(port int) (bool, error) {
	index, err := GetPortIndex(port)
	if err != nil {
		return false, err
	}
	if index > -1 && index < locals.ACTIVE_TRANSACTION_LIMIT {
		ports[index][1] = 0
		return true, nil
	}
	return false, log.WRONG_PORT_INDEX("FreePort()")
}