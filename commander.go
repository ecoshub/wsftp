package main

import (
	"fmt"
	"net"
	"strconv"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/websocket"
	"github.com/eco9999/jparse"
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
		json := jparse.Parse(<- commandChan)
		event, result := json.GetString("event")
		if !result {continue}
		if event != ""{
			switch event{
			case "creq":
				if activeTransaction < ACTIVETRANSACTIONLIMIT {
					dir, result := json.GetString("dir")
					if !result {continue}
					mac, result := json.GetString("mac")
					if !result {continue}
					uuid, result := json.GetString("uuid")
					if !result {continue}
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
				dir, result := json.GetString("dir")
				if !result {continue}
				dest, result := json.GetString("dest")
				if !result {continue}
				uuid, result := json.GetString("uuid")
				if !result {continue}
				mac, result := json.GetString("mac")
				if !result {continue}
				username := hs.GetUsername(mac)
				ip := hs.GetIP(mac)
				if newPort == -1{
					cmd.TransmitData(myIP,SRLISTENPORT,`{"event":"info","content":"Active transaction full"}`)
					cmd.SendReject(ip, mac, dir, uuid, username)
				}else{
					portIDMap[newPort] = uuid
					go com.ReceiveFile(ip, mac, username, newPort, uuid, &(ports[index][1]))
					cmd.SendAccept(ip, mac, dir, dest, username, uuid, newPort)
				}
			case "crej":
				mac, result := json.GetString("mac")
				if !result {continue}
				dir, result := json.GetString("dir")
				if !result {continue}
				uuid, result := json.GetString("uuid")
				if !result {continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendReject(ip, mac, dir, uuid, username)
			case "cmsg":
				mac, result := json.GetString("mac")
				if !result {continue}
				msg, result := json.GetString("msg")
				if !result {continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendMessage(ip, mac, username, msg)
			case "racp":
				dir, result := json.GetString("dir")
				if !result {continue}
				dest, result := json.GetString("destination")
				if !result {continue}
				uuid, result := json.GetString("uuid")
				if !result {continue}
				mac, result := json.GetString("mac")
				if !result {continue}
				ip, result := json.GetString("ip")
				if !result {continue}
				port, result := json.GetString("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				index := getPortIndex(intPort)
				username := hs.GetUsername(mac)
				setPortBusy(intPort)
				go com.SendFile(ip, mac, username, intPort, uuid, dir, dest, &(ports[index][1]))
			case "dprg":
				port, result := json.GetString("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "fprg":
				port, result := json.GetString("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "kprg":
				port, result := json.GetString("port")
				if !result {continue}
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