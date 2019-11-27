package main

import (
	"fmt"
	"net"
	"strconv"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/websocket"
	hs "wsftp/hs"
	com "wsftp/tcpcom"
	utils "wsftp/utils"
	cmd "wsftp/cmd"
	router "wsftp/router"
	json "wsftp/json"
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
	ACTIVEDOWNLOADLIMIT int = 25
	ACTIVEUPLOADLIMIT int = 25
)

var (
	activeUpload int = 0
	activeDownload int = 0
	ports = make([][]int, ACTIVEDOWNLOADLIMIT)
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
		receive := string(<- commandChan)
		rec := json.JTM(receive)
		stat := rec["cmd"][0]
		if stat != ""{
			switch stat{
			case "creq":
				if activeUpload < ACTIVEUPLOADLIMIT {
					dir := rec["dir"][0]
					// of receiver
					mac := rec["mac"][0]
					ip := hs.GetIP(mac)
					username := hs.GetUsername(mac)
					cmd.SendRequest(ip, dir, mac, username)
					activeUpload++
				}else{
					cmd.SendMsg(myIP,SRLISTENPORT,`{"stat":"info","content":"activeUploadFull"}`)
				}
			case "cacp":
				index := allocatePort()
				newPort := ports[index][0]
				if newPort == -1{
					cmd.SendMsg(myIP,SRLISTENPORT,`{"stat":"info","content":"activeDownloadFull"}`)
					cmd.SendReject(rec["ip"][0],rec["dir"][0])
				}else{
					go com.ReceiveFile(rec["ip"][0], newPort, &(ports[index][1]))
					cmd.SendAccept(rec["ip"][0], rec["dir"][0], rec["dest"][0], newPort)
				}
			case "crej":
				cmd.SendReject(rec["ip"][0],rec["dir"][0])
			case "cmsg":
				cmd.SendMessage(rec["ip"][0],rec["to"][0], rec["msg"][0])
			case "racp":
				intPort, _ := strconv.Atoi(rec["port"][0])
				index := getPortIndex(intPort)
				setPortBusy(intPort)
				go com.SendFile(rec["ip"][0], intPort, rec["dir"][0], rec["destination"][0], &(ports[index][1]))
			case "dprg":
				intPort, _ := strconv.Atoi(rec["port"][0])
				freePort(intPort)
			case "fprg":
				intPort, _ := strconv.Atoi(rec["port"][0])
				freePort(intPort)
			case "kprg":
				intPort, _ := strconv.Atoi(rec["port"][0])
				freePort(intPort)
			case "reshs":
				hs.Restart()
			}
		}else{
			fmt.Println("Wrong command")
		}
	}
}

func allocatePort() int{
	for i := 0 ; i < ACTIVEDOWNLOADLIMIT ; i++ {
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
	for i := 0 ; i < ACTIVEDOWNLOADLIMIT ; i++ {
		if portCheck(STARTPORT - i){
			ports[i] = []int{STARTPORT - i, 0}
		}
	}
}

func getPortIndex(port int) int{
	for i := 0 ; i < ACTIVEDOWNLOADLIMIT ; i++ {
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