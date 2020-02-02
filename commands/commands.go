package commands

import (
	"github.com/ecoshub/jint"
	"github.com/ecoshub/penman"
	"net"
	"strconv"
	"wsftp/tools"
)

const (
	WS_COMMANDER_LISTEN_PORT    string = "9999"
	WS_SEND_RECEIVE_LISTEN_PORT string = "10001"
	WS_MESSAGE_LISTEN_PORT      string = "10002"

	ERROR_ADDRESS_RESOLVING string = "Commands: TCP IP resolve error."
	ERROR_CONNECTION_FAILED string = "Commands: TCP Connection error."
)

var (
	MY_USERNAME string = tools.MY_USERNAME
	MY_IP       string = tools.MY_IP
	MY_MAC      string = tools.MY_MAC

	REQUEST_SCHEME *jint.Scheme = jint.MakeScheme("event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "fileSize", "contentType", "uuid")
	CANCEL_SCHEME  *jint.Scheme = jint.MakeScheme("event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "fileSize", "contentType", "uuid")
	ACCEPT_SCHEME  *jint.Scheme = jint.MakeScheme("event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "dest", "port", "uuid", "contentType")
	REJECT_SCHEME  *jint.Scheme = jint.MakeScheme("event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "uuid", "cause", "contentType")
	MESSAGE_SCHEME *jint.Scheme = jint.MakeScheme("event", "username", "nick", "mac", "content", "contentType")

	WARNING_FILE_NOT_FOUND []byte = jint.MakeJson([]string{"event", "content"}, []string{"info", "File not found or size is zero"})
)

func SendRequest(ip, dir, mac, username, nick, uuid string) {
	dir = penman.PreProcess(dir)
	fileSize := tools.GetFileSize(dir)
	fileName := tools.GetFileName(dir)
	fileType := tools.GetFileExt(fileName)
	if fileSize == 0 {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, WARNING_FILE_NOT_FOUND)
		return
	}
	rreq := REQUEST_SCHEME.MakeJson("rreq", MY_USERNAME, tools.MY_NICK, MY_IP, MY_MAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	sreq := REQUEST_SCHEME.MakeJson("sreq", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	freq := REQUEST_SCHEME.MakeJson("freq", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	if sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, rreq) {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, sreq)
	} else {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, freq)
	}
}

func SendCancel(ip, dir, mac, username, nick, uuid string) {
	dir = penman.PreProcess(dir)
	fileSize := tools.GetFileSize(dir)
	fileName := tools.GetFileName(dir)
	fileType := tools.GetFileExt(fileName)
	if fileSize == 0 {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, WARNING_FILE_NOT_FOUND)
		return
	}
	rcncl := CANCEL_SCHEME.MakeJson("rcncl", MY_USERNAME, tools.MY_NICK, MY_IP, MY_MAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	scncl := CANCEL_SCHEME.MakeJson("scncl", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	fcncl := CANCEL_SCHEME.MakeJson("fcncl", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid)
	if sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, rcncl) {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, scncl)
	} else {
		sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, fcncl)
	}
}

func SendAccept(ip, mac, dir, dest, username, nick, uuid string, port int) {
	fileName := tools.GetFileName(dir)
	fileType := tools.GetFileExt(fileName)
	racp := ACCEPT_SCHEME.MakeJson("racp", MY_USERNAME, tools.MY_NICK, MY_IP, MY_MAC, dir, fileName, fileType, dest, strconv.Itoa(port), uuid, "file")
	sacp := ACCEPT_SCHEME.MakeJson("sacp", username, nick, ip, mac, dir, fileName, fileType, dest, strconv.Itoa(port), uuid, "file")
	facp := ACCEPT_SCHEME.MakeJson("facp", username, nick, ip, mac, dir, fileName, fileType, dest, strconv.Itoa(port), uuid, "file")
	
	if sendCore(ip, WS_COMMANDER_LISTEN_PORT, racp) {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, sacp)
		sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, sacp)
	} else {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, facp)
		sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, facp)
	}
}

func SendReject(ip, mac, dir, uuid, username, nick, cause string) {
	fileName := tools.GetFileName(dir)
	fileType := tools.GetFileExt(fileName)
	rrej := ACCEPT_SCHEME.MakeJson("rrej", MY_USERNAME, tools.MY_NICK, MY_IP, MY_MAC, dir, fileName, fileType, uuid, cause, "file")
	srej := ACCEPT_SCHEME.MakeJson("srej", username, nick, ip, mac, dir, fileName, fileType, uuid, cause, "file")
	frej := ACCEPT_SCHEME.MakeJson("frej", username, nick, ip, mac, dir, fileName, fileType, uuid, cause, "file")
	if sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, rrej) {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, srej)
	} else {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, frej)
	}
}

func SendMessage(ip, mac, username, nick, msg string) {
	rmsg := MESSAGE_SCHEME.MakeJson("rmsg", MY_USERNAME, tools.MY_NICK, MY_MAC, msg, "text")
	smsg := MESSAGE_SCHEME.MakeJson("smsg", username, nick, mac, msg, "text")
	fmsg := MESSAGE_SCHEME.MakeJson("fmsg", username, nick, mac, msg, "text")
	if sendCore(ip, WS_SEND_RECEIVE_LISTEN_PORT, rmsg) {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, smsg)
	} else {
		sendCore(tools.MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, fmsg)
	}
}

func sendCore(ip, port string, data []byte) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip+":"+port)
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
