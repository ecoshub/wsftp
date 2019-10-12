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

var startPort int = 9996
var mainListen string = "9997"
// 9998 reserved for handshake to handshake com.
var mainListenInt int = 9999
// 10000 reserved for handshake
var srListen int = 10001
var msgListen int = 10002
// 10003 reserved for ws sr
// 10004 reserved for ws msg

var endPoint = "/cmd"
var activeDownload int = 25
var activeUploadLimit int = 25
var activeUpload int = 0
var ports = make([][]int, activeDownload)
var myIP string = utils.GetInterfaceIP().String()
var myUserName string = utils.GetUsername()
var asd int = 31

var commandChan = make(chan []byte, 1)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main(){
	initPorts()
	// startup
	go hs.Start()
	go router.StartRouting()

	http.HandleFunc(endPoint, handleConn)
	go listen()
	go manage()
	err := http.ListenAndServe(":" + mainListen, nil)
	fmt.Println("Command Server shutdown unexpectedly!", err)
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
    }
    conn.Close()
	commandChan <- msg
	return true
}

func listen(){
	for {
		receive(mainListenInt)
	}
}

func handleConn(w http.ResponseWriter, r *http.Request){
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer ws.Close()
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			ws.Close()
			fmt.Println("Main Command Connection Closed: ", err)
			break
		}
		commandChan <- msg
	}
}

func manage(){
	for {
		receive := string(<- commandChan)
		rec := json.JTM(receive)
		stat := rec["stat"][0]
		if stat != ""{
			switch stat{
			case "creq":
				if activeUpload < activeUploadLimit {
					cmd.SendRequest(rec["ip"][0],rec["dir"][0])
					activeUpload++
				}else{
					cmd.SendMsg(myIP,srListen,`{"stat":"info","content":"activeUploadFull"}`)
				}
			case "cacp":
				index := allocatePort()
				newPort := ports[index][0]
				if newPort == -1{
					cmd.SendMsg(myIP,srListen,`{"stat":"info","content":"activeDownloadFull"}`)
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
				go com.SendFile(rec["ip"][0], intPort, rec["dir"][0], rec["destination"][0], &(ports[index][1]))
				setPortBusy(intPort)
			case "dprg":
				port := rec["port"][0]
				intPort, _ := strconv.Atoi(port)
				freePort(intPort)
			case "fprg":
				ip := rec["ip"][0]
				intPort, _ := strconv.Atoi(rec["port"][0])
				cmd.SendMsg(ip, srListen, receive)
				
				rec["ip"] = []string{myIP}
				newRec := json.MTJ(rec)
				cmd.SendMsg(myIP, srListen, newRec)

				freePort(intPort)
			case "ncon":
				ip := rec["ip"][0]
				rec["from"] = []string{myIP}
				intPort, _ := strconv.Atoi(rec["port"][0])
				cmd.SendMsg(ip, srListen, receive)
				
				rec["from"] = []string{ip}
				rec["ip"] = []string{myIP}
				newRec := json.MTJ(rec)
				cmd.SendMsg(myIP, srListen, newRec)

				freePort(intPort)

			case "kprg":
				intPort, _ := strconv.Atoi(rec["port"][0])
				freePort(intPort)
			}
		}else{
			fmt.Println("Wrong command")
		}
	}
}

func allocatePort() int{
	for i := 0 ; i < activeDownload ; i++ {
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
	for i := 0 ; i < activeDownload ; i++ {
		if portCheck(startPort - i){
			ports[i] = []int{startPort - i, 0}
		}
	}
}

func getPortIndex(port int) int{
	for i := 0 ; i < activeDownload ; i++ {
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