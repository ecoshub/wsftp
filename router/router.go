package router

import (
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net"
	"net/http"
	"wsftp/tools"
)

const (
	TCP_SEND_RECEIVE_LISTEN_PORT           string = "10001"
	TCP_MESSAGE_LISTEN_PORT                string = "10002"
	WS_SEND_RECEIVE_LISTEN_PORT            string = "10003"
	WS_MESSAGE_LISTEN_PORT                 string = "10004"
	MESSAGE_END_POINT                      string = "/msg"
	SEND_RECEIVE_END_POINT                 string = "/sr"
	ERROR_WS_CONNECTION_MESSAGE            string = "Router: Fatal Error: Message websocket connection error."
	ERROR_WS_CONNECTION_SEND_RECEIVE       string = "Router: Fatal Error: Send/Receive websocket connection error."
	ERROR_UNEXPECTED_SHUTDOWN_MESSAGE      string = "Router: Fatal Error: Message websocket shutdown unexpectedly."
	ERROR_UNEXPECTED_SHUTDOWN_SEND_RECEIVE string = "Router: Fatal Error: Send/Receive websocket shutdown unexpectedly."
	ERROR_TCP_LISTEN_FAILED                string = "Router: TCP Listen error."
	ERROR_TCP_READ                         string = "Router: TCP read error."
	ERROR_LISTEN_ACCECPT_FAILED            string = "Router: TCP Listen accept error."
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
		tools.StdoutHandle("fatal", ERROR_WS_CONNECTION_MESSAGE, err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-messageChan))
	}
}

func handleSr(w http.ResponseWriter, r *http.Request) {
	ws, err := upgraderSR.Upgrade(w, r, nil)
	if err != nil {
		tools.StdoutHandle("fatal", ERROR_WS_CONNECTION_SEND_RECEIVE, err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-srChan))
	}
}

func StartMessageChan() {
	http.HandleFunc(MESSAGE_END_POINT, handleMessage)
	err := http.ListenAndServe(":"+WS_MESSAGE_LISTEN_PORT, nil)
	tools.StdoutHandle("fatal", ERROR_UNEXPECTED_SHUTDOWN_MESSAGE, err)
}

func StartSrChan() {
	http.HandleFunc(SEND_RECEIVE_END_POINT, handleSr)
	err := http.ListenAndServe(":"+WS_SEND_RECEIVE_LISTEN_PORT, nil)
	tools.StdoutHandle("fatal", ERROR_UNEXPECTED_SHUTDOWN_SEND_RECEIVE, err)
}

func msgListen(ch chan<- []byte) {
	listener, err := net.Listen("tcp", ":"+TCP_MESSAGE_LISTEN_PORT)
	if err != nil {
		tools.StdoutHandle("error", ERROR_TCP_LISTEN_FAILED, err)
	} else {
		defer listener.Close()
	}
	conn, err := listener.Accept()
	if err != nil {
		tools.StdoutHandle("error", ERROR_LISTEN_ACCECPT_FAILED, err)
	} else {
		defer conn.Close()
		msg, err := ioutil.ReadAll(conn)
		if err != nil {
			tools.StdoutHandle("error", ERROR_TCP_READ, err)
		} else {
			ch <- msg
		}
	}
}

func srListen(ch chan<- []byte) {
	listener, err := net.Listen("tcp", ":"+TCP_SEND_RECEIVE_LISTEN_PORT)
	if err != nil {
		tools.StdoutHandle("error", ERROR_TCP_LISTEN_FAILED, err)
	} else {
		defer listener.Close()
	}
	conn, err := listener.Accept()
	if err != nil {
		tools.StdoutHandle("error", ERROR_LISTEN_ACCECPT_FAILED, err)
	} else {
		defer conn.Close()
		msg, err := ioutil.ReadAll(conn)
		if err != nil {
			tools.StdoutHandle("error", ERROR_TCP_READ, err)
		} else {
			ch <- msg
		}
	}
}
