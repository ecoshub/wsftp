package main

import "fmt"
import "wsftp/tools"
import "wsftp/port_tools"
import "wsftp/locals"


func main(){
	fmt.Println("hello")
	fmt.Println(tools.Msg())
	fmt.Println(tools.Message)
	fmt.Println(locals.TCP_TRANSECTION_START_PORT,
	locals.WS_COMMANDER_LISTEN_PORT,
	locals.UDP_HANDSHAKE_LISTEN_PORT,
	locals.WS_HANDSHAKE_LISTEN_PORT,
	locals.WS_SEND_RECEIVE_LISTEN_PORT,
	locals.WS_MESSAGE_LISTEN_PORT)

	port := port_tools.AllocatePort()
	fmt.Println(port)
	fmt.Println(port_tools.MainPortCheck())
}