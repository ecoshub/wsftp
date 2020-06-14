package ports

import "strconv"
import "net"
import "errors"

var (
	// ports
	TCP_TRANSECTION_START_PORT   int = 9996
	WS_COMMANDER_LISTEN_PORT     int = 9997
	UDP_HANDSHAKE_LISTEN_PORT    int = 9998
	TCP_COMMANDER_LISTEN_PORT    int = 9999
	WS_HANDSHAKE_LISTEN_PORT     int = 10000
	TCP_SEND_RECEIVE_LISTEN_PORT int = 10001
	TCP_MESSAGE_LISTEN_PORT      int = 10002
	WS_SEND_RECEIVE_LISTEN_PORT  int = 10003
	WS_MESSAGE_LISTEN_PORT       int = 10004

	ACTIVE_TRANSACTION_LIMIT int = 25
	ActiveTransaction        int = 0
	availablePorts               = make([][]int, ACTIVE_TRANSACTION_LIMIT)
	uuids                        = make([]string, ACTIVE_TRANSACTION_LIMIT)

	ERROR_MAIN_PORT_BUSSY string = "PortTools: The ports that required for the program to work properly is busy. Please close other program/programs that using this ports. Port range is [9997:10002]"
	ERROR_PORT_INDEX_GET  string = "PortTools: Port index out of range. at GetPortIndex()"
	ERROR_PORT_INDEX_SET  string = "PortTools: Port index out of range. at SetPortBusy()"
	ERROR_PORT_INDEX_FREE string = "PortTools: Port index out of range. at FreePort()"
)

func init() {
	// port initializing
	for i := 0; i < ACTIVE_TRANSACTION_LIMIT; i++ {
		if portCheck(TCP_TRANSECTION_START_PORT - i) {
			availablePorts[i] = []int{TCP_TRANSECTION_START_PORT - i, 0}
		}
	}
}

func AllocatePort() int {
	for i := 0; i < ACTIVE_TRANSACTION_LIMIT; i++ {
		if availablePorts[i][1] == 0 && portCheck(availablePorts[i][0]) {
			availablePorts[i][1] = 1
			return i
		}
	}
	return -1
}

func MainPortCheck() error {
	result := portCheck(TCP_TRANSECTION_START_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(WS_COMMANDER_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(UDP_HANDSHAKE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(TCP_COMMANDER_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(WS_HANDSHAKE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(TCP_SEND_RECEIVE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(TCP_MESSAGE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(WS_SEND_RECEIVE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	result = result && portCheck(WS_MESSAGE_LISTEN_PORT)
	if !result {
		return errors.New(ERROR_MAIN_PORT_BUSSY)
	}
	return nil
}

func portCheck(port int) bool {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

func GetPortIndex(port int) (int, error) {
	for i := 0; i < ACTIVE_TRANSACTION_LIMIT; i++ {
		if availablePorts[i][0] == port {
			return i, nil
		}
	}
	return -1, errors.New(ERROR_PORT_INDEX_GET)
}

func GetPort(index int) int{
	return availablePorts[index][0]
}

func SetPortBusy(port int) error {
	index, err := GetPortIndex(port)
	if err != nil {
		return err
	}
	if index > -1 && index < ACTIVE_TRANSACTION_LIMIT {
		availablePorts[index][1] = 1
		return nil
	}
	return errors.New(ERROR_PORT_INDEX_SET)
}

func FreePort(port int) error {
	index, err := GetPortIndex(port)
	if err != nil {
		return err
	}
	if index > -1 && index < ACTIVE_TRANSACTION_LIMIT {
		availablePorts[index][1] = 0
		return nil
	}
	return errors.New(ERROR_PORT_INDEX_FREE)
}

func GetControl(index int) *int {
	return &(availablePorts[index][1])
}

func AllocateUUID(uuid string) {
	uuids = append(uuids, uuid)
}

func ClearUUID(uuid string) {
	newuuids := make([]string, 0, len(uuids))
	for _, u := range uuids {
		if u != uuid {
			newuuids = append(newuuids, u)
		}
	}
	uuids = newuuids
}

func HasUUID(uuid string) bool {
	for _, u := range uuids {
		if u == uuid {
			return true
		}
	}
	return false
}
