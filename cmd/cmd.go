package commands

import (
	"fmt"
	jint "github.com/ecoshub/jint"
	penman "github.com/ecoshub/penman"
	"net"
	"strconv"
	utils "wsftp/utils"
)

const (
	// 9996 reserverd for transfer start port
	// 9997 reserverd ws commander comminication.
	// 9998 reserverd tcp handshake comminication.
	MAINLISTEN int = 9999
	// 10000 reserved for handshake
	SRLISTEN  int = 10001
	MSGLISTEN int = 10002
	// 10003 reserved for ws sr
	// 10004 reserved for ws msg
)

var (
	myIP       string = utils.GetInterfaceIP().String()
	myUsername string = utils.GetUsername()
	myNick     string = utils.GetNick()
	myMAC      string = utils.GetEthMac()
)

func SendRequest(ip, dir, mac, username, nick, uuid string) {

	dir = penman.PreProcess(dir)
	fileSize := utils.GetFileSize(dir)
	fileName := utils.GetFileName(dir)
	fileType := utils.GetFileExt(fileName)

	if fileSize == 0 {
		fileNotFound := jint.MakeJson([]string{"event", "content"}, []string{"info", "File not found or size is zero"})
		TransmitDataByte(myIP, SRLISTEN, fileNotFound)
		return
	}

	data := jint.Scheme([]string{"event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "fileSize", "contentType", "uuid"})

	dataForReceiver := data.MakeJson([]string{"none", myUsername, myNick, myIP, myMAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid})
	dataForMe := data.MakeJson([]string{"none", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid})

	rreq, _ := jint.SetString(dataForReceiver, "rreq", "event")
	sreq, _ := jint.SetString(dataForMe, "sreq", "event")
	freq, _ := jint.SetString(dataForMe, "freq", "event")

	if TransmitDataByte(ip, SRLISTEN, rreq) {
		TransmitDataByte(myIP, SRLISTEN, sreq)
	} else {
		TransmitDataByte(myIP, SRLISTEN, freq)
	}
}

func SendCancel(ip, dir, mac, username, nick, uuid string) {

	dir = penman.PreProcess(dir)
	fileSize := utils.GetFileSize(dir)
	fileName := utils.GetFileName(dir)
	fileType := utils.GetFileExt(fileName)

	data := jint.Scheme([]string{"event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "fileSize", "contentType", "uuid"})

	dataForReceiver := data.MakeJson([]string{"none", myUsername, myNick, myIP, myMAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid})
	dataForMe := data.MakeJson([]string{"none", username, nick, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10), "file", uuid})

	rcncl, _ := jint.SetString(dataForReceiver, "rcncl", "event")
	scncl, _ := jint.SetString(dataForMe, "scncl", "event")
	fcncl, _ := jint.SetString(dataForMe, "fcncl", "event")

	TransmitDataByte(ip, SRLISTEN, rcncl)
	if TransmitDataByte(ip, SRLISTEN, rcncl) {
		TransmitDataByte(myIP, SRLISTEN, scncl)
	} else {
		TransmitDataByte(myIP, SRLISTEN, fcncl)
	}
}

func SendAccept(ip, mac, dir, dest, username, nick, uuid string, port int) {

	fileName := utils.GetFileName(dir)
	fileType := utils.GetFileExt(fileName)

	data := jint.Scheme([]string{"event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "dest", "port", "uuid", "contentType"})

	dataForReceiver := data.MakeJson([]string{"none", myUsername, myNick, myIP, myMAC, dir, fileName, fileType, dest, strconv.Itoa(port), uuid, "file"})
	dataForMe := data.MakeJson([]string{"none", username, nick, ip, mac, dir, fileName, fileType, dest, strconv.Itoa(port), uuid, "file"})

	racp, _ := jint.SetString(dataForReceiver, "racp", "event")
	sacp, _ := jint.SetString(dataForMe, "sacp", "event")
	facp, _ := jint.SetString(dataForMe, "facp", "event")

	if TransmitDataByte(ip, MAINLISTEN, racp) {
		TransmitDataByte(myIP, SRLISTEN, sacp)
		TransmitDataByte(ip, SRLISTEN, racp)
	} else {
		TransmitDataByte(ip, SRLISTEN, facp)
		TransmitDataByte(myIP, SRLISTEN, facp)
	}
}

func SendReject(ip, mac, dir, uuid, username, nick, cause string) {

	fileName := utils.GetFileName(dir)
	fileType := utils.GetFileExt(fileName)

	data := jint.Scheme([]string{"event", "username", "nick", "ip", "mac", "dir", "fileName", "fileType", "uuid", "cause", "contentType"})

	dataForReceiver := data.MakeJson([]string{"none", myUsername, myNick, myIP, myMAC, dir, fileName, fileType, uuid, cause, "file"})
	dataForMe := data.MakeJson([]string{username, nick, ip, mac, dir, fileName, fileType, uuid, cause, "file"})

	rrej, _ := jint.SetString(dataForReceiver, "rrej", "event")
	srej, _ := jint.SetString(dataForMe, "srej", "event")
	frej, _ := jint.SetString(dataForMe, "frej", "event")

	if TransmitDataByte(ip, SRLISTEN, rrej) {
		TransmitDataByte(myIP, SRLISTEN, srej)
	} else {
		TransmitDataByte(myIP, SRLISTEN, frej)
	}
}

func SendMessage(ip, mac, username, nick, msg string) {

	data := jint.Scheme([]string{"event", "mac", "username", "nick", "content", "contentType"})

	dataForReceiver := data.MakeJson([]string{"none", myMAC, myUsername, myNick, msg, "text"})
	dataForMe := data.MakeJson([]string{"none", mac, username, nick, msg, "text"})

	rmsg, _ := jint.SetString(dataForReceiver, "rmsg", "event")
	smsg, _ := jint.SetString(dataForMe, "smsg", "event")
	fmsg, _ := jint.SetString(dataForMe, "fmsg", "event")

	if TransmitDataByte(ip, SRLISTEN, rmsg) {
		TransmitDataByte(myIP, SRLISTEN, smsg)
	} else {
		TransmitDataByte(myIP, SRLISTEN, fmsg)
	}
}

func TransmitDataByte(ip string, port int, data []byte) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Address resolving error (Inner)", err)
		return false
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Connection Fail (Inner)", err)
		return false
	} else {
		conn.Write(data)
		conn.Close()
		return true
	}
}

func TransmitData(ip string, port int, data string) bool {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ip+":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Address resolving error (Inner)", err)
		return false
	}
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		fmt.Println("Connection Fail (Inner)", err)
		return false
	} else {
		conn.Write([]byte(data))
		conn.Close()
		return true
	}
}
