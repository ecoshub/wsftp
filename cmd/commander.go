package main

import (
	"github.com/ecoshub/jint"
	"github.com/ecoshub/penman"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"wsftp/commands"
	"wsftp/handshake"
	"wsftp/tools"
	"wsftp/transaction"
)

const (
	WS_COMMANDER_LISTEN_PORT    int = 9997
	TCP_COMMANDER_LISTEN_PORT   int = 9999
	WS_SEND_RECEIVE_LISTEN_PORT int = 10001

	LOG_START                   string = "Main: Websocket listen started."
	ERROR_GET_IP                string = "Main: Fatal Error: IP resolve error. Commander closing."
	ERROR_GET_MAC               string = "Main: Fatal Error: MAC resolve error. Commander closing."
	ERROR_ADDRESS_RESOLVING     string = "Main: TCP IP resolve error."
	ERROR_CONNECTION_FAILED     string = "Main: TCP Connection error."
	ERROR_TCP_LISTEN_FAILED     string = "Main: TCP Listen error."
	ERROR_TCP_READ              string = "Main: TCP read error."
	ERROR_LISTEN_ACCECPT_FAILED string = "Main: TCP Listen accept error."
	ERROR_PORTS_BUSY            string = "Main: Fatal Error: Ports Busy. Commander closing."
	ERROR_WS_CONNECTION         string = "Main: Fatal Error: Websocket connection error. Commander closing."
	ERROR_WS_READ               string = "Main: Fatal Error: Websocket read error Pleas refresh client wensocket. Commander closing."
	ERROR_UNEXPECTED_SHUTDOWN   string = "Main: Fatal Error: Websocket shutdown unexpectedly. Commander closing."
	INFO_WS_CONNECTION          string = "Main: Websocket connected."
	INFO_FOLDER                 string = "Main: Folder transaction not suppoted."
	INFO_TRANSACTION_FULL       string = "Main: Active transaction full."
	INFO_WRONG_COMMAND          string = "Main: Wrong command."
	INFO_NULL_EVENT             string = "Main: 'event' key can not be null."
	ERROR_GET_PORT              string = "Main: GetPort error."
	ERROR_SET_PORT              string = "Main: SetPort error."
	ERROR_FREE_PORT             string = "Main: FreePort error."
	ERROR_JSON_PARSE            string = "Main: JSON parse error. Probably missing key."
	COMMANDER_END_POINT         string = "/cmd"
)

var (
	MY_SCHEME   *jint.Scheme = jint.MakeScheme("event", "username", "mac", "ip", "nick")
	loopControl bool         = true
	MY_IP       string
	MY_MAC      string
	commandChan = make(chan []byte, 1)
	upgrader    = websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: false,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	var err error
	MY_IP, err = tools.GetInterfaceIP()
	if err != nil {
		tools.StdoutHandle("fatal", ERROR_GET_IP, err)
		return
	}
	MY_MAC, err = tools.GetMac()
	if err != nil {
		tools.StdoutHandle("fatal", ERROR_GET_MAC, err)
		return
	}
	err = tools.MainPortCheck()
	if err != nil {
		tools.StdoutHandle("fatal", ERROR_GET_MAC, err)
		return
	}
	go handshake.Start()
	go listenTCP()
	tools.StdoutHandle("log", LOG_START, nil)
	http.HandleFunc(COMMANDER_END_POINT, handleConn)
	err = http.ListenAndServe(":"+strconv.Itoa(WS_COMMANDER_LISTEN_PORT), nil)
	tools.StdoutHandle("fatal", ERROR_UNEXPECTED_SHUTDOWN, err)
}

func manage() {
	for {
		json := <-commandChan
		event, err := jint.GetString(json, "event")
		if err != nil {
			parseErrorHandle(err, "event")
			continue
		}
		if event != "" {
			switch event {
			case "actv":
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, jint.MakeJson([]string{"event", "total", "active"}, []string{"actv", strconv.Itoa(tools.ACTIVE_TRANSACTION_LIMIT), strconv.Itoa(tools.ActiveTransaction)}))
			case "my":
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, MY_SCHEME.MakeJson(tools.MY_USERNAME, tools.MY_MAC, MY_IP, tools.GetNick()))
			case "creq":
				dir, err := jint.GetString(json, "dir")
				if err != nil {
					parseErrorHandle(err, "dir")
					continue
				}
				if tools.ActiveTransaction < tools.ACTIVE_TRANSACTION_LIMIT {
					mac, err := jint.GetString(json, "mac")
					if err != nil {
						parseErrorHandle(err, "mac")
						continue
					}
					uuid, err := jint.GetString(json, "uuid")
					if err != nil {
						parseErrorHandle(err, "uuid")
						continue
					}
					ip, err := jint.GetString(json, "ip")
					if err != nil {
						parseErrorHandle(err, "ip")
						continue
					}
					username, err := jint.GetString(json, "username")
					if err != nil {
						parseErrorHandle(err, "username")
						continue
					}
					nick, err := jint.GetString(json, "nick")
					if err != nil {
						parseErrorHandle(err, "nick")
						continue
					}
					if penman.IsDir(dir) {
						sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_FOLDER, nil))
						continue
					} else {
						commands.SendRequest(ip, dir, mac, username, nick, uuid)
						tools.ActiveTransaction++
						continue
					}
				} else {
					sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_TRANSACTION_FULL, nil))
					continue
				}
			case "cacp":
				dir, err := jint.GetString(json, "dir")
				if err != nil {
					parseErrorHandle(err, "dir")
					continue
				}
				if tools.ActiveTransaction < tools.ACTIVE_TRANSACTION_LIMIT {
					dest, err := jint.GetString(json, "dest")
					if err != nil {
						parseErrorHandle(err, "dest")
						continue
					}
					uuid, err := jint.GetString(json, "uuid")
					if err != nil {
						parseErrorHandle(err, "uuid")
						continue
					}
					mac, err := jint.GetString(json, "mac")
					if err != nil {
						parseErrorHandle(err, "mac")
						continue
					}
					ip, err := jint.GetString(json, "ip")
					if err != nil {
						parseErrorHandle(err, "ip")
						continue
					}
					username, err := jint.GetString(json, "username")
					if err != nil {
						parseErrorHandle(err, "username")
						continue
					}
					nick, err := jint.GetString(json, "nick")
					if err != nil {
						parseErrorHandle(err, "nick")
						continue
					}
					index := tools.AllocatePort()
					if index == -1 {
						sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_TRANSACTION_FULL, nil))
						commands.SendReject(ip, mac, dir, uuid, username, nick, "full")
						continue
					}
					newPort := tools.Ports[index][0]
					// portIDMap[newPort] = uuid
					go transaction.ReceiveFile(ip, mac, username, nick, newPort, uuid, &(tools.Ports[index][1]))
					commands.SendAccept(ip, mac, dir, dest, username, nick, uuid, newPort)
					tools.ActiveTransaction++
				} else {
					sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_TRANSACTION_FULL, nil))
				}
			case "crej":
				mac, err := jint.GetString(json, "mac")
				if err != nil {
					parseErrorHandle(err, "mac")
					continue
				}
				dir, err := jint.GetString(json, "dir")
				if err != nil {
					parseErrorHandle(err, "dir")
					continue
				}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {
					parseErrorHandle(err, "uuid")
					continue
				}
				ip, err := jint.GetString(json, "ip")
				if err != nil {
					parseErrorHandle(err, "ip")
					continue
				}
				username, err := jint.GetString(json, "username")
				if err != nil {
					parseErrorHandle(err, "username")
					continue
				}
				nick, err := jint.GetString(json, "nick")
				if err != nil {
					parseErrorHandle(err, "nick")
					continue
				}
				commands.SendReject(ip, mac, dir, uuid, username, nick, "standart")
			case "cmsg":
				mac, err := jint.GetString(json, "mac")
				if err != nil {
					parseErrorHandle(err, "mac")
					continue
				}
				msg, err := jint.GetString(json, "msg")
				if err != nil {
					parseErrorHandle(err, "msg")
					continue
				}
				ip, err := jint.GetString(json, "ip")
				if err != nil {
					parseErrorHandle(err, "ip")
					continue
				}
				username, err := jint.GetString(json, "username")
				if err != nil {
					parseErrorHandle(err, "username")
					continue
				}
				nick, err := jint.GetString(json, "nick")
				if err != nil {
					parseErrorHandle(err, "nick")
					continue
				}
				commands.SendMessage(ip, mac, username, nick, msg)
			case "racp":
				dir, err := jint.GetString(json, "dir")
				if err != nil {
					parseErrorHandle(err, "dir")
					continue
				}
				dest, err := jint.GetString(json, "dest")
				if err != nil {
					parseErrorHandle(err, "dest")
					continue
				}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {
					parseErrorHandle(err, "uuid")
					continue
				}
				mac, err := jint.GetString(json, "mac")
				if err != nil {
					parseErrorHandle(err, "mac")
					continue
				}
				ip, err := jint.GetString(json, "ip")
				if err != nil {
					parseErrorHandle(err, "ip")
					continue
				}
				port, err := jint.GetString(json, "port")
				if err != nil {
					parseErrorHandle(err, "port")
					continue
				}
				username, err := jint.GetString(json, "username")
				if err != nil {
					parseErrorHandle(err, "username")
					continue
				}
				nick, err := jint.GetString(json, "nick")
				if err != nil {
					parseErrorHandle(err, "nick")
					continue
				}
				intPort, _ := strconv.Atoi(port)
				index, err := tools.GetPortIndex(intPort)
				if err != nil {
					tools.StdoutHandle("info", ERROR_GET_PORT, err)
					continue
				}
				err = tools.SetPortBusy(intPort)
				if err != nil {
					tools.StdoutHandle("info", ERROR_SET_PORT, err)
					continue
				}
				go transaction.SendFile(ip, mac, username, nick, intPort, uuid, dir, dest, &(tools.Ports[index][1]))
			case "cncl":
				dir, err := jint.GetString(json, "dir")
				if err != nil {
					parseErrorHandle(err, "dir")
					continue
				}
				uuid, err := jint.GetString(json, "uuid")
				if err != nil {
					parseErrorHandle(err, "uuid")
					continue
				}
				mac, err := jint.GetString(json, "mac")
				if err != nil {
					parseErrorHandle(err, "mac")
					continue
				}
				ip, err := jint.GetString(json, "ip")
				if err != nil {
					parseErrorHandle(err, "ip")
					continue
				}
				username, err := jint.GetString(json, "username")
				if err != nil {
					parseErrorHandle(err, "username")
					continue
				}
				nick, err := jint.GetString(json, "nick")
				if err != nil {
					parseErrorHandle(err, "nick")
					continue
				}
				commands.SendCancel(ip, dir, mac, username, nick, uuid)
			case "dprg":
				port, err := jint.GetString(json, "port")
				if err != nil {
					parseErrorHandle(err, "port")
					continue
				}
				intPort, _ := strconv.Atoi(port)
				err = tools.FreePort(intPort)
				if err != nil {
					tools.StdoutHandle("info", ERROR_FREE_PORT, err)
					continue
				}
				tools.ActiveTransaction--

			case "fprg":
				port, err := jint.GetString(json, "port")
				if err != nil {
					parseErrorHandle(err, "port")
					continue
				}
				intPort, _ := strconv.Atoi(port)
				err = tools.FreePort(intPort)
				if err != nil {
					tools.StdoutHandle("info", ERROR_FREE_PORT, err)
					continue
				}
				tools.ActiveTransaction--
			case "kprg":
				port, err := jint.GetString(json, "port")
				if err != nil {
					parseErrorHandle(err, "port")
					continue
				}
				intPort, _ := strconv.Atoi(port)
				err = tools.FreePort(intPort)
				if err != nil {
					tools.StdoutHandle("info", ERROR_FREE_PORT, err)
					continue
				}
			case "rshs":
				handshake.Restart()
			default:
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_WRONG_COMMAND, nil))
			}
		} else {
			sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, tools.LOG_SCHEME.MakeJson("info", INFO_NULL_EVENT, nil))
		}
	}
}

func handleConn(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		tools.StdoutHandle("fatal", ERROR_WS_CONNECTION, err)
		return
	} else {
		tools.StdoutHandle("info", INFO_WS_CONNECTION, nil)
		for {
			_, msg, err := ws.ReadMessage()
			if err != nil {
				tools.StdoutHandle("fatal", ERROR_WS_READ, err)
				return
			}
			commandChan <- msg
		}
	}
}

func receiveTCP() bool {
	addr := MY_IP + ":" + strconv.Itoa(TCP_COMMANDER_LISTEN_PORT)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		tools.StdoutHandle("error", ERROR_ADDRESS_RESOLVING, err)
		commandChan <- []byte{0}
		return false
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		tools.StdoutHandle("error", ERROR_TCP_LISTEN_FAILED, err)
		listener.Close()
		commandChan <- []byte{0}
		return false
	}
	defer listener.Close()
	conn, err := listener.Accept()
	if err != nil {
		tools.StdoutHandle("error", ERROR_LISTEN_ACCECPT_FAILED, err)
		return false
	}
	msg, err := ioutil.ReadAll(conn)
	if err != nil {
		tools.StdoutHandle("error", ERROR_TCP_READ, err)
	} else {
		commandChan <- msg
	}
	return true
}

func listenTCP() {
	for loopControl {
		receiveTCP()
	}
}

func parseErrorHandle(err error, key string) {
	tools.StdoutHandle("info", ERROR_JSON_PARSE+" '"+key+"'", err)
}

func sendCore(ip string, port int, data []byte) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		tools.StdoutHandle("warning", ERROR_ADDRESS_RESOLVING, err)
		return false
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_CONNECTION_FAILED, err)
		return false
	} else {
		conn.Write(data)
		conn.Close()
		return true
	}
}
