package commands

import (
	"fmt"
	"net"
	"strconv"
	utils "wsftp/utils"
)

const (
    // 9996 reserverd for transfer start port
    // 9997 reserverd ws commander comminication.
    // 9998 reserverd tcp handshake comminication.
    mainListen int = 9999
    // 10000 reserved for handshake
    srListen int = 10001
    msgListen int = 10002
    // 10003 reserved for ws sr
    // 10004 reserved for ws msg
)

var myIP string = utils.GetInterfaceIP().String()

func SendRequest(ip, dir string){
    fileSize := utils.GetFileSize(dir)
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
	username := utils.GetUsername()
    data := fmt.Sprintf(`"username":"%v","ip":"%v","dir":"%v","fileName":"%v","fileType":"%v","fileSize":"%v","contentType":"file"}`,
     username, myIP, dir, fileName, fileType, strconv.FormatInt(fileSize, 10))
    rreq := `{"stat":"rreq",` + data
    sreq := `{"stat":"sreq",` + data
    freq := `{"stat":"freq",` + data
    res := SendMsg(ip, srListen, rreq)
    if res {
        SendMsg(myIP, srListen, sreq)
    }else{
        SendMsg(myIP, srListen, freq)
    }
}

func SendAccept(ip, dir , dest string, port int){
    username := utils.GetUsername()
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
    data := fmt.Sprintf(`"username":"%v","ip":"%v","dir":"%v","fileName":"%v","fileType":"%v","destination":"%v","port":"%v","contentType":"file"}`,
        username, myIP, dir, fileName, fileType, dest, strconv.Itoa(port))
    racp := `{"stat":"racp",` + data
    sacp := `{"stat":"sacp",` + data
    facp := `{"stat":"facp",` + data
    
    res := SendMsg(ip, mainListen, racp)
    if res {
        SendMsg(myIP, srListen, sacp)
        SendMsg(ip, srListen, racp)
    }else{
        SendMsg(ip, srListen, facp)
        SendMsg(myIP, srListen, facp)
    }
}

func SendReject(ip, dir string){
    username := utils.GetUsername()
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
    data := fmt.Sprintf(`"username":"%v","ip":"%v","dir":"%v","fileName":"%v","fileType":"%v","contentType":"file"}`,
     username, myIP, dir, fileName, fileType)
    rrej := `{"stat":"rrej",` + data
    srej := `{"stat":"srej",` + data
    frej := `{"stat":"frej",` + data
    res := SendMsg(ip, srListen, rrej)
    if res {
    	SendMsg(myIP, srListen, srej)
    }else{
    	SendMsg(myIP, srListen, frej)
    }
}

func SendMessage(ip, to, msg string){
    username := utils.GetUsername()
    data := fmt.Sprintf(`"person":"%v","content":"%v","contentType":"text"}`, username, msg)
    data2 := fmt.Sprintf(`"person":"%v","content":"%v","contentType":"text"}`, to, msg)
    rmsg := `{"stat":"rmsg",` + data
    smsg := `{"stat":"smsg",` + data2
    fmsg := `{"stat":"fmsg",` + data2
    res := SendMsg(ip,msgListen, rmsg)
    if res {
        SendMsg(myIP, msgListen, smsg)
    }else{
        SendMsg(myIP, msgListen, fmsg)
    }
}

func SendMsg(ip string, port int, msg string) bool{
    strPort := strconv.Itoa(port)
    addr := ip + ":" + strPort
    tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
        fmt.Println("Address resolving error (Inner)", err)
        return false
    }
    conn, err := net.DialTCP("tcp", nil, tcpAddr)
    if err != nil {
        fmt.Println("Connection Fail (Inner)", err)
        return false
    }else{
        conn.Write([]byte(msg))
        conn.Close()
        return true
    }
}