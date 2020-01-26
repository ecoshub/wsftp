package main

import (
	"fmt"
	"net"
	"strconv"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/websocket"
	"github.com/ecoshub/jint"
	hs "wsftp/hs"
	com "wsftp/tcpcom"
	utils "wsftp/utils"
	cmd "wsftp/cmd"
	router "wsftp/router"
)

const (
	// ports
	STARTPORT int = 9996
	MAINLISTENPORTWS int = 9997
	// 9998 reserved for handshake to handshake com.
	MAINLISTENPORT int = 9999
	// 10000 reserved for handshake to frontend ws com.
	SRLISTENPORT int = 10001
	MSGLISTENPORT int = 10002
	// 10003 reserved for sr to frontend ws com.
	// 10004 reserved for msg to frontend ws com.

	// websocket settings & limits
	ENDPOINT = "/cmd"
	ACTIVETRANSACTIONLIMIT int = 25
)

var (
	activeTransaction int = 0
	ports = make([][]int, ACTIVETRANSACTIONLIMIT)
	portIDMap = make(map[int]string, ACTIVETRANSACTIONLIMIT)
	myIP string = utils.GetInterfaceIP().String()
	commandChan = make(chan []byte, 1)

	upgrader = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main(){
	initPorts()
	go hs.Start()
	go router.StartRouting()
	go listen()
	go manage()
	http.HandleFunc(ENDPOINT, handleConn)
	err := http.ListenAndServe(":" + strconv.Itoa(MAINLISTENPORTWS), nil)
	fmt.Println("Commander shutdown unexpectedly!", err)
}

func receive(port int) bool {
    strPort := strconv.Itoa(port)
    addr := myIP + ":" + strPort
    tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
        fmt.Println("Address resolving error (Inner)",err)
		commandChan <- []byte{0}
        return false
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        fmt.Println("Listen Error (Inner)", err)
		listener.Close()
		commandChan <- []byte{0}
		return false

    }else{
        defer listener.Close()
    }
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Listen Accept Error (Inner) ", err)
		return false
    }
    msg, err :=  ioutil.ReadAll(conn)
    if err != nil {
        fmt.Println("Message Read Error (Inner)", err)
    }else{
	    conn.Close()
		commandChan <- msg
    }
	return true
}

func listen(){
	for {
		receive(MAINLISTENPORT)
	}
}

func handleConn(w http.ResponseWriter, r *http.Request){
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}else{
		fmt.Println("Connection Establish")
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				fmt.Println("Main Command Connection Closed: ", err)
				fmt.Println("Waiting For Another Connection...")
				fmt.Println("If You Are Using 'Commander Tester' Please Restart This Program For synchronization")
				break
			}
			commandChan <- msg
		}
	}
}

func manage(){
	for {
		json := <- commandChan
		event, err := jint.GetString(json, "event")
		if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'event'\n", err));continue}
		if event != ""{
			switch event{
			case "my":
				cmd.TransmitData(myIP, SRLISTENPORT,fmt.Sprintf(`{"event":"my","username":"%v","mac":"%v","ip":"%v"}`, utils.GetCustomUsername(), utils.GetEthMac(), myIP))
			case "creq":
				if activeTransaction < ACTIVETRANSACTIONLIMIT {
					dir, err := jint.GetString(json, "dir")
					if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'dir'\n", err));continue}
					mac, err := jint.GetString(json, "mac")
					if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'mac'\n", err));continue}
					uuid, err := jint.GetString(json, "uuid")
					if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'uuid'\n", err));continue}
					ip := hs.GetIP(mac)
					username := hs.GetUsername(mac)
					cmd.SendRequest(ip, dir, mac, username, uuid)
					activeTransaction++
				}else{
					cmd.TransmitData(myIP, SRLISTENPORT,`{"event":"info","content":"Active transaction full"}`)
				}
			case "cacp":
				index := allocatePort()
				newPort := ports[index][0]
				dir, err := jint.GetString(json, "dir")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'dir'\n", err));continue}
				dest, err := jint.GetString(json, "dest")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'dest'\n", err));continue}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'uuid'\n", err));continue}
				mac, err := jint.GetString(json, "mac")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'mac'\n", err));continue}
				username := hs.GetUsername(mac)
				ip := hs.GetIP(mac)
				if newPort == -1{
					cmd.TransmitData(myIP,SRLISTENPORT,`{"event":"info","content":"Active transaction full"}`)
					cmd.SendReject(ip, mac, dir, uuid, username)
					// send info message to receiver to
				}else{
					portIDMap[newPort] = uuid
					go com.ReceiveFile(ip, mac, username, newPort, uuid, &(ports[index][1]))
					cmd.SendAccept(ip, mac, dir, dest, username, uuid, newPort)
				}
			case "crej":
				mac, err := jint.GetString(json, "mac")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'mac'\n", err));continue}
				dir, err := jint.GetString(json, "dir")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'dir'\n", err));continue}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'uuid'\n", err));continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendReject(ip, mac, dir, uuid, username)
			case "cmsg":
				mac, err := jint.GetString(json, "mac")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'mac'\n", err));continue}
				msg, err := jint.GetString(json, "msg")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'msg'\n", err));continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendMessage(ip, mac, username, msg)
			case "racp":
				dir, err := jint.GetString(json, "dir")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'dir'\n", err));continue}
				dest, err := jint.GetString(json, "destination")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'destination'\n", err));continue}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'uuid'\n", err));continue}
				mac, err := jint.GetString(json, "mac")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'mac'\n", err));continue}
				ip, err := jint.GetString(json, "ip")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'ip'\n", err));continue}
				port, err := jint.GetString(json, "port")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'port'\n", err));continue}
				intPort, _ := strconv.Atoi(port)
				index := getPortIndex(intPort)
				username := hs.GetUsername(mac)
				setPortBusy(intPort)
				go com.SendFile(ip, mac, username, intPort, uuid, dir, dest, &(ports[index][1]))
			case "dprg":
				port, err := jint.GetString(json, "port")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'port'\n", err));continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "fprg":
				port, err := jint.GetString(json, "port")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'port'\n", err));continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "kprg":
				port, err := jint.GetString(json, "port")
				if err != nil {cmd.TransmitData(myIP, SRLISTENPORT, fmt.Sprintf("(Commander JSON Parse Error). ERROR: %v, KEY: 'port'\n", err));continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
			case "rshs":
				hs.Restart()
			}
		}else{
			cmd.TransmitData(myIP,SRLISTENPORT,`{"event":"info","content":"Wrong command"}`)
		}
	}
}

func allocatePort() int{
	for i := 0 ; i < ACTIVETRANSACTIONLIMIT ; i++ {
		if ports[i][1] == 0  && portCheck(ports[i][0]){
			ports[i][1] = 1
			return i
		}
	}
	return -1
}

func portCheck(port int) bool{
    strPort := strconv.Itoa(port)
    listener, err := net.Listen("tcp", ":" + strPort)
    if err != nil {
       return false
    }
    listener.Close()
    return true
}

func initPorts(){
	for i := 0 ; i < ACTIVETRANSACTIONLIMIT ; i++ {
		if portCheck(STARTPORT - i){
			ports[i] = []int{STARTPORT - i, 0}
		}
	}
}

func getPortIndex(port int) int{
	for i := 0 ; i < ACTIVETRANSACTIONLIMIT ; i++ {
		if ports[i][0] == port {
			return i
		}
	}
	fmt.Println("Fatal error port index out of range", port)
	return -1
}

func setPortBusy(port int) {
	index := getPortIndex(port)
	ports[index][1] = 1
}

func freePort(port int){
	index := getPortIndex(port)
	ports[index][1] = 0 
}