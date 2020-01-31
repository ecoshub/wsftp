package routers

import (
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net"
	"net/http"
)

const (
	// 9996 reserverd for transfer start port
	// 9997 reserverd ws commander comminication.
	// 9998 reserverd tcp handshake comminication.
	// 9999 reserverd tcp commander comminication.
	// 10000 reserverd ws handshake comminication.
	SRTCPPORT       string = "10001"
	MSGTCPPORT      string = "10002"
	SRWSPORT        string = "10003"
	MSGWSPORT       string = "10004"
	MESSAGEENDPOINT string = "/msg"
	SRENDPOINT      string = "/sr"
)

var (
	messageChan    = make(chan []byte, 1)
	srChan         = make(chan []byte, 1)
	srReceiveChan  = make(chan []byte, 1)
	msgReceiveChan = make(chan []byte, 1)

	upgraderSR = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	upgraderMSG = websocket.Upgrader{
		ReadBufferSize:    4096,
		WriteBufferSize:   4096,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func StartRouting() {
	go StartMessageChan()
	go StartSrChan()
	go startMSGListen()
	go startSRListen()

}

func startSRListen() {
	for {
		srListen(srReceiveChan)
		tempsr := <-srReceiveChan
		srChan <- tempsr
	}
}

func startMSGListen() {
	for {
		msgListen(msgReceiveChan)
		tempmsg := <-msgReceiveChan
		messageChan <- tempmsg
	}
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
	ws, err := upgraderMSG.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Route message channel websocket connection error: ", err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-messageChan))
	}
}

func handleSr(w http.ResponseWriter, r *http.Request) {
	ws, err := upgraderSR.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Route sr channel websocket connection error: ", err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-srChan))
	}
}

func StartMessageChan() {
	http.HandleFunc(MESSAGEENDPOINT, handleMessage)
	err := http.ListenAndServe(":"+MSGWSPORT, nil)
	fmt.Println("Server shutdown unexpectedly!", err)
}

func StartSrChan() {
	http.HandleFunc(SRENDPOINT, handleSr)
	err := http.ListenAndServe(":"+SRWSPORT, nil)
	fmt.Println("Server shutdown unexpectedly!", err)
}

func msgListen(ch chan<- []byte) {
	listener, err := net.Listen("tcp", ":"+MSGTCPPORT)
	if err != nil {
		fmt.Println("Listen Error (Router)")
	} else {
		defer listener.Close()
	}
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Listen Accept Error (Router)")
	} else {
		defer conn.Close()
		msg, err := ioutil.ReadAll(conn)
		if err != nil {
			fmt.Println("Message Read Error (Router)")
		} else {
			ch <- msg
		}
	}
}

func srListen(ch chan<- []byte) {
	listener, err := net.Listen("tcp", ":"+SRTCPPORT)
	if err != nil {
		fmt.Println("Listen Error (Router)")
	} else {
		defer listener.Close()
	}
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Listen Accept Error (Router)")
	} else {
		defer conn.Close()
		msg, err := ioutil.ReadAll(conn)
		if err != nil {
			fmt.Println("Message Read Error (Router)")
		} else {
			ch <- msg
		}
	}
}
