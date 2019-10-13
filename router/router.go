package routers

import (
	"fmt"
	"net"
	"net/http"
	"io/ioutil"
	"github.com/gorilla/websocket"
	utils "wsftp/utils"
)

// 9996 reserverd for transfer start port
// 9997 reserverd ws commander comminication.
// 9998 reserverd tcp handshake comminication.
// 9999 reserverd tcp commander comminication.
// 10000 reserverd ws handshake comminication.
var srTCPPort string = "10001"
var msgTCPPort string = "10002"
var srWSPort string = "10003"
var msgWSPort string = "10004"

var myIP string = utils.GetInterfaceIP().String()
var messageEndPoint string = "/msg"
var srEndPoint string = "/sr"

var messageChan = make(chan []byte, 1)
var srChan = make(chan []byte, 1)
var srReceiveChan = make(chan []byte, 1)
var msgReceiveChan = make(chan []byte, 1)

var upgraderSR = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var upgraderMSG = websocket.Upgrader{
	ReadBufferSize:    4096,
	WriteBufferSize:   4096,
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func StartRouting(){
	go StartMessageChan()
	go StartSrChan()
	go startMSGListen()
	go startSRListen()

}

func startSRListen(){
	for {
		srListen(srReceiveChan)
		tempsr := <- srReceiveChan
		srChan <- tempsr
	}
}

func startMSGListen(){
	for {
		msgListen(msgReceiveChan)
		tempmsg := <- msgReceiveChan
		messageChan <- tempmsg
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request){
	ws, err := upgraderMSG.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-messageChan))
	}
}

func handleSr(w http.ResponseWriter, r *http.Request){
	ws, err := upgraderSR.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-srChan))
	}
}

func StartMessageChan(){
	http.HandleFunc(messageEndPoint, handleMessage)
	err := http.ListenAndServe(":" + msgWSPort, nil)
	fmt.Println("Server shutdown unexpectedly!", err)
}

func StartSrChan(){
	http.HandleFunc(srEndPoint, handleSr)
	err := http.ListenAndServe(":" + srWSPort, nil)
	fmt.Println("Server shutdown unexpectedly!", err)
}

func msgListen(ch chan<- []byte){
	listener, err := net.Listen("tcp", ":" + msgTCPPort)
    if err != nil {
        fmt.Println("Listen Error (Router)")
    }else{
    	defer listener.Close()
    }
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Listen Accept Error (Router)")
    }else{
    	defer conn.Close()
	    msg, err :=  ioutil.ReadAll(conn)
	    if err != nil {
	        fmt.Println("Message Read Error (Router)")
	    }else{
	    	ch <- msg
	    }
    }
}

func srListen(ch chan<- []byte){
	listener, err := net.Listen("tcp", ":" + srTCPPort)
    if err != nil {
        fmt.Println("Listen Error (Router)")
    }else{
    	defer listener.Close()
    }
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Listen Accept Error (Router)")
    }else{
    	defer conn.Close()
	    msg, err :=  ioutil.ReadAll(conn)
	    if err != nil {
	        fmt.Println("Message Read Error (Router)")
	    }else{
	    	ch <- msg
	    }
    }
}