package handshake

import (
	"github.com/ecoshub/jint"
	"github.com/gorilla/websocket"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"wsftp/tools"
)

const (
	// connections
	LOOP_CONTROL_LIMIT        int    = 100
	UDP_SIGNAL_REPEAT         int    = 5
	BROADCAST_LISTEN_IP       string = "0.0.0.0"
	UDP_HANDSHAKE_LISTEN_PORT string = "9998"
	WS_HANDSHAKE_LISTEN_PORT  string = "10000"
	HANDSHAKE_END_POINT       string = "/hs"

	// log
	ERROR_BROADCAST_CONNECTION string = "Handshake: Broadcast UDP connection error."
	LOG_START                  string = "Handshake: Websocket listen started."
	ERROR_CONNECTION           string = "Handshake: Websocket connection error."
	ERROR_CLOSE                string = "Handshake: Websocket listen server shutdown unexpectedly."
	ERROR_EMPTY_UDP            string = "Handshake: Empty UDP message has arrived."
	ERROR_BAD_UDP              string = "Handshake: Bad UDP message has arrived."
	ERROR_UDP_READ             string = "Handshake: UDP read error."
	ERROR_UDP_SIGNAL           string = "Handshake: UDP send error."
	ERROR_JSON_PARSE           string = "Handshake: JSON parse error. Probably missing key."
)

var (
	// json tools
	HANDSHAKESCHEME *jint.Scheme = jint.MakeScheme("event", "ip", "username", "nick", "mac")
	ONLINE_MESSAGE  []byte       = HANDSHAKESCHEME.MakeJson("online", tools.MY_IP, tools.MY_USERNAME, tools.MY_NICK, tools.MY_MAC)
	OFFLINE_MESSAGE []byte       = HANDSHAKESCHEME.MakeJson("offline", tools.MY_IP, tools.MY_USERNAME, tools.MY_NICK, tools.MY_MAC)

	// base variablees
	receiveControl   bool     = true
	loopControl      bool     = true
	MACList          []string = make([]string, 0, 1024)
	onlineList                = make(map[string][]string, 256)
	innerMessageChan          = make(chan []byte, 1)
	signalChannel             = make(chan os.Signal, 1)

	// websocket
	upgraderHS = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func handleConn(w http.ResponseWriter, r *http.Request) {
	ws, err := upgraderHS.Upgrade(w, r, nil)
	if err != nil {
		tools.StdoutHandle("error", ERROR_CONNECTION, err)
	}
	defer ws.Close()
	for loopControl {
		ws.WriteMessage(1, []byte(<-innerMessageChan))
	}
}

func Start() {
	go activity()
	http.HandleFunc(HANDSHAKE_END_POINT, handleConn)
	tools.StdoutHandle("log", LOG_START, nil)
	err := http.ListenAndServe(":"+WS_HANDSHAKE_LISTEN_PORT, nil)
	tools.StdoutHandle("error", ERROR_CLOSE, err)
}

func Restart() {
	sendMessage(OFFLINE_MESSAGE)
	tools.MY_NICK = tools.GetNick()
	ONLINE_MESSAGE = HANDSHAKESCHEME.MakeJson("online", tools.MY_IP, tools.MY_USERNAME, tools.MY_NICK, tools.MY_MAC)
	OFFLINE_MESSAGE = HANDSHAKESCHEME.MakeJson("offline", tools.MY_IP, tools.MY_USERNAME, tools.MY_NICK, tools.MY_MAC)
	MACList = make([]string, 0, 1024)
	onlineList = make(map[string][]string, 128)
	sendMessage(ONLINE_MESSAGE)
}

func activity() {
	sendMessage(ONLINE_MESSAGE)
	receiveChan := make(chan []byte, 1)
	done := make(chan bool, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChannel
		sendMessageValidation(OFFLINE_MESSAGE, done)
		receiveControl = false
		<-done
		os.Exit(0)
	}()
	for receiveControl {
		go listenUDP(receiveChan)
		message := <-receiveChan
		if len(message) < 2 {
			if len(message) == 1 {
				if message[0] == 0 {
					tools.StdoutHandle("warning", ERROR_BAD_UDP, nil)
					continue
				}
			}
			if len(message) == 0 {
				tools.StdoutHandle("warning", ERROR_EMPTY_UDP, nil)
				continue
			}
		}
		tempStatus, err := jint.GetString(message, "event")
		if err != nil {
			tools.StdoutHandle("warning", ERROR_JSON_PARSE+" 'event'", err)
			continue
		}
		tempIP, err := jint.GetString(message, "ip")
		if err != nil {
			tools.StdoutHandle("warning", ERROR_JSON_PARSE+" 'ip'", err)
			continue
		}
		tempUsername, err := jint.GetString(message, "username")
		if err != nil {
			tools.StdoutHandle("warning", ERROR_JSON_PARSE+" 'username'", err)
			continue
		}
		tempMAC, err := jint.GetString(message, "mac")
		if err != nil {
			tools.StdoutHandle("warning", ERROR_JSON_PARSE+" 'mac'", err)
			continue
		}
		tempNick, err := jint.GetString(message, "nick")
		if err != nil {
			tools.StdoutHandle("warning", ERROR_JSON_PARSE+" 'nick'", err)
			continue
		}
		receiving_message := HANDSHAKESCHEME.MakeJson(tempStatus, tempIP, tempUsername, tempNick, tempMAC)
		if tempMAC != tools.MY_MAC {
			if !hasThis(MACList, tempMAC) && tempStatus == "online" && tempUsername != tools.MY_USERNAME {
				onlineList[tempMAC] = []string{tempUsername, tempIP}
				MACList = append(MACList, tempMAC)
				innerMessageChan <- []byte(receiving_message)
				sendMessage(ONLINE_MESSAGE)
			}
			if hasThis(MACList, tempMAC) && tempStatus == "offline" && tempUsername != tools.MY_USERNAME {
				MACList = removeFromList(MACList, tempMAC)
				delete(onlineList, tempMAC)
				innerMessageChan <- []byte(receiving_message)
			}
		}
	}
}

func listenUDP(ch chan<- []byte) {
	buff := make([]byte, 1024)
	pack, err := net.ListenPacket("udp", BROADCAST_LISTEN_IP+":"+UDP_HANDSHAKE_LISTEN_PORT)
	if err != nil {
		tools.StdoutHandle("error", ERROR_BROADCAST_CONNECTION, err)
		ch <- []byte{0}
	}
	defer pack.Close()
	n, _, err := pack.ReadFrom(buff)
	if err != nil {
		tools.StdoutHandle("error", ERROR_UDP_READ, err)
		ch <- []byte{0}
	}
	ch <- buff[:n]
}

func sendMessage(data []byte) {
	validation := make(chan bool, 1)
	count := 0
	valid := false
	for !valid && count < UDP_SIGNAL_REPEAT {
		sendCore(data, validation)
		valid = <-validation
		count++
	}
	if !valid {
		tools.StdoutHandle("error", ERROR_UDP_SIGNAL, nil)
	}
}

func sendMessageValidation(data []byte, ch chan<- bool) {
	validation := make(chan bool, 1)
	count := 0
	valid := false
	for !valid && count < UDP_SIGNAL_REPEAT {
		sendCore(data, validation)
		valid = <-validation
		count++
	}
	if !valid {
		tools.StdoutHandle("error", ERROR_UDP_SIGNAL, nil)
		ch <- false
	}
	ch <- true
}

func sendCore(data []byte, ch chan<- bool) {
	conn, err := net.Dial("udp", tools.BROADCAST_IP+":"+UDP_HANDSHAKE_LISTEN_PORT)
	if err != nil {
		ch <- false
	} else {
		defer conn.Close()
		conn.Write(data)
		ch <- true
	}
}

func hasThis(list []string, el string) bool {
	for _, v := range list {
		if v == el {
			return true
		}
	}
	return false
}

func removeFromList(list []string, el string) []string {
	lenl := len(list)
	if lenl < 2 {
		return nil
	}
	newList := make([]string, lenl-1, 1024)
	count := 0
	for _, v := range list {
		if v != el {
			newList[count] = v
			count++
		}
	}
	return newList
}
