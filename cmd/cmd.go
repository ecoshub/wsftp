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
    MAINLISTEN int = 9999
    // 10000 reserved for handshake
    SRLISTEN int = 10001
    MSGLISTEN int = 10002
    // 10003 reserved for ws sr
    // 10004 reserved for ws msg
)

var myIP string = utils.GetInterfaceIP().String()

func SendRequest(ip, dir , mac string){
    fileSize := utils.GetFileSize(dir)
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
	username := utils.GetUsername()
    myMAC := utils.GetEthMac()
    data := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","fileSize":"%v","contentType":"file"}`,
     username, myIP, myMAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10))
    rreq := `{"stat":"rreq",` + data

    data = fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","fileSize":"%v","contentType":"file"}`,
     username, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10))
    sreq := `{"stat":"sreq",` + data
    freq := `{"stat":"freq",` + data
    res := SendMsg(ip, SRLISTEN, rreq)
    if res {
        SendMsg(myIP, SRLISTEN, sreq)
    }else{
        SendMsg(myIP, SRLISTEN, freq)
    }
}

func SendAccept(ip, dir , dest string, port int){
    username := utils.GetUsername()
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
    data := fmt.Sprintf(`"username":"%v","ip":"%v","dir":"%v","fileName":"%v","fileType":"%v","destination":"%v","port":"%v","contentType":"file"}`,
        username, ip, dir, fileName, fileType, dest, strconv.Itoa(port))
    racp := `{"stat":"racp",` + data
    sacp := `{"stat":"sacp",` + data
    facp := `{"stat":"facp",` + data
    
    res := SendMsg(ip, MAINLISTEN, racp)
    if res {
        SendMsg(myIP, SRLISTEN, sacp)
        SendMsg(ip, SRLISTEN, racp)
    }else{
        SendMsg(ip, SRLISTEN, facp)
        SendMsg(myIP, SRLISTEN, facp)
    }
}

func SendReject(ip, dir string){
    username := utils.GetUsername()
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
    data := fmt.Sprintf(`"username":"%v","ip":"%v","dir":"%v","fileName":"%v","fileType":"%v","contentType":"file"}`,
     username, ip, dir, fileName, fileType)
    rrej := `{"stat":"rrej",` + data
    srej := `{"stat":"srej",` + data
    frej := `{"stat":"frej",` + data
    res := SendMsg(ip, SRLISTEN, rrej)
    if res {
    	SendMsg(myIP, SRLISTEN, srej)
    }else{
    	SendMsg(myIP, SRLISTEN, frej)
    }
}

func SendMessage(ip, to, msg string){
    username := utils.GetUsername()
    data := fmt.Sprintf(`"person":"%v","content":"%v","contentType":"text"}`, username, msg)
    data2 := fmt.Sprintf(`"person":"%v","content":"%v","contentType":"text"}`, to, msg)
    rmsg := `{"stat":"rmsg",` + data
    smsg := `{"stat":"smsg",` + data2
    fmsg := `{"stat":"fmsg",` + data2
    res := SendMsg(ip,MSGLISTEN, rmsg)
    if res {
        SendMsg(myIP, MSGLISTEN, smsg)
    }else{
        SendMsg(myIP, MSGLISTEN, fmsg)
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