package main

import (
	"fmt"
	// "net/http"
	// "wsftp/tools"
	"wsftp/tools"
	// "wsftp/locals"
	// "github.com/gorilla/websocket"
	// "wsftp/handshake"
	"wsftp/commands"
)

// var (
// 	commandChan              = make(chan []byte, 1)
// 	upgrader = websocket.Upgrader{
// 		ReadBufferSize:    1024,
// 		WriteBufferSize:   1024,
// 		EnableCompression: false,
// 		CheckOrigin: func(r *http.Request) bool {
// 			return true
// 		},
// 	}
// )

/*

	check ports
	check ip
	check mac

*/

func main(){
	// fmt.Println(tools.Msg())
	// fmt.Println(tools.Message)
	// fmt.Println(locals.TCP_TRANSECTION_START_PORT,
	// locals.WS_COMMANDER_LISTEN_PORT,
	// locals.WS_HANDSHAKE_LISTEN_PORT,
	// locals.WS_SEND_RECEIVE_LISTEN_PORT,
	// locals.WS_MESSAGE_LISTEN_PORT)

	err := tools.MainPortCheck()
	if err != nil {
		fmt.Println(err)
		return
	}
	// fmt.Println(tools.SEPARATOR)
	// fmt.Println(tools.MY_USERNAME)
	// fmt.Println(tools.MY_NICK)
	// fmt.Println(tools.MY_IP)
	// fmt.Println(tools.MY_MAC)
	// fmt.Println(tools.SETTINGS_DIR)
	// fmt.Println(tools.BROADCAST_IP)
	// handshake.Start()

	// fmt.Println(commands.WARNING_FILE_NOT_FOUND)
	// ip, mac, dir, dest, username, nick, uuid string, port int
	// commands.SendAccept("192.168.1.105","bc:ae:c5:13:84:f9","/home/ecomain/Desktop/tbbt.avi","/home/ecomain/Desktop","ecolab","ecolab","15378351ec8432a35c3513242d", 9996)
	commands.SendMessage("192.168.1.105","bc:ae:c5:13:84:f9","ecolab","ecolab","hello gardaş!")
}