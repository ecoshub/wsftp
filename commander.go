package main

import (
	"fmt"
	"net"
	"strconv"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/websocket"
	parser "github.com/eco9999/jparse"
	hs "wsftp/hs"
	com "wsftp/tcpcom"
	utils "wsftp/utils"
	cmd "wsftp/cmd"
	router "wsftp/router"
	rw "wsftp/rw"
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
	myUserName string = utils.GetUsername()
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
	// test
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
		receivedJSONCommand, done := parser.JSONParser(string(<- commandChan))
		if !done{continue}
		event, result := receivedJSONCommand.Get("event")
		if !result {continue}
		if event != ""{
			switch event{
			case "save":
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				input, result := receivedJSONCommand.Get("input")
				if !result {continue}
				username, result := receivedJSONCommand.Get("username")
				if !result {continue}
				content, result := receivedJSONCommand.Get("content")
				if !result {continue}
				rw.SaveLog(username, mac, input, content)
				cmd.TransmitData(myIP, SRLISTENPORT, `{"event":"info","content":"saved"}`)
			case "get":
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				username, result := receivedJSONCommand.Get("username")
				if !result {continue}
				start, result := receivedJSONCommand.Get("start")
				if !result {continue}
				end, result := receivedJSONCommand.Get("end")
				if !result {continue}
				content, result := receivedJSONCommand.Get("content")
				if !result {continue}
				startN, _ := strconv.Atoi(start)
				endN, _ := strconv.Atoi(end)
				log := rw.GetLog(username, mac, content, startN, endN)
				str := fmt.Sprintf(`{"event":"log","mac":"%v","username":"%v","data":[%v]}`, mac, username, log)
				cmd.TransmitData(myIP, SRLISTENPORT, str)
			case "creq":
				if activeTransaction < ACTIVETRANSACTIONLIMIT {
					dir, result := receivedJSONCommand.Get("dir")
					if !result {continue}
					mac, result := receivedJSONCommand.Get("mac")
					if !result {continue}
					ip := hs.GetIP(mac)
					username := hs.GetUsername(mac)
					cmd.SendRequest(ip, dir, mac, username)
					activeTransaction++
				}else{
					cmd.TransmitData(myIP, SRLISTENPORT,`{"event":"info","content":"Active transaction full"}`)
				}
			case "cacp":
				index := allocatePort()
				newPort := ports[index][0]
				dir, result := receivedJSONCommand.Get("dir")
				if !result {continue}
				dest, result := receivedJSONCommand.Get("dest")
				if !result {continue}
				id, result := receivedJSONCommand.Get("id")
				if !result {continue}
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				username := hs.GetUsername(mac)
				ip := hs.GetIP(mac)
				if newPort == -1{
					cmd.TransmitData(myIP,SRLISTENPORT,`{"event":"info","content":"Active transaction full"}`)
					cmd.SendReject(ip, mac, dir, username)
				}else{
					portIDMap[newPort] = id
					go com.ReceiveFile(ip, mac, username, newPort, id, &(ports[index][1]))
					cmd.SendAccept(ip, mac, dir, dest, username, id, newPort)
				}
			case "crej":
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				dir, result := receivedJSONCommand.Get("dir")
				if !result {continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendReject(ip, mac, dir, username)
			case "cmsg":
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				msg, result := receivedJSONCommand.Get("msg")
				if !result {continue}
				ip := hs.GetIP(mac)
				username := hs.GetUsername(mac)
				cmd.SendMessage(ip, mac, username, msg)
			case "racp":
				dir, result := receivedJSONCommand.Get("dir")
				if !result {continue}
				dest, result := receivedJSONCommand.Get("destination")
				if !result {continue}
				id, result := receivedJSONCommand.Get("id")
				if !result {continue}
				mac, result := receivedJSONCommand.Get("mac")
				if !result {continue}
				ip, result := receivedJSONCommand.Get("ip")
				if !result {continue}
				port, result := receivedJSONCommand.Get("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				index := getPortIndex(intPort)
				username := hs.GetUsername(mac)
				setPortBusy(intPort)
				go com.SendFile(ip, mac, username, intPort, id, dir, dest, &(ports[index][1]))
			case "dprg":
				port, result := receivedJSONCommand.Get("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "fprg":
				port, result := receivedJSONCommand.Get("port")
				if !result {continue}
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
				activeTransaction--
			case "kprg":
				port, result := receivedJSONCommand.Get("port")
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