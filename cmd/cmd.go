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

var (
    myIP string = utils.GetInterfaceIP().String()
    myUsername string = utils.GetUsername()
    myMAC string = utils.GetEthMac()
)

func SendRequest(ip, dir , mac , username string){
    fileSize := utils.GetFileSize(dir)
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)
    
    dataToSend := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","fileSize":"%v","contentType":"file"}`,
     myUsername, myIP, myMAC, dir, fileName, fileType, strconv.FormatInt(fileSize, 10))
    dataToMe := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","fileSize":"%v","contentType":"file"}`,
     username, ip, mac, dir, fileName, fileType, strconv.FormatInt(fileSize, 10))

    rreq := `{"cmd":"rreq",` + dataToSend
    sreq := `{"cmd":"sreq",` + dataToMe
    freq := `{"cmd":"freq",` + dataToMe
    res := TransmitData(ip, SRLISTEN, rreq)
    if res {
        TransmitData(myIP, SRLISTEN, sreq)
    }else{
        TransmitData(myIP, SRLISTEN, freq)
    }
}

func SendAccept(ip, mac, dir, dest, username, id string, port int){
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)

    dataToSend := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","destination":"%v","port":"%v","id":"%v","contentType":"file"}`,
        myUsername, myIP, myMAC, dir, fileName, fileType, dest, strconv.Itoa(port), id)
    dataToMe := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","destination":"%v","port":"%v","id":"%v","contentType":"file"}`,
        username, ip, mac, dir, fileName, fileType, dest, strconv.Itoa(port), id)

    racp := `{"cmd":"racp",` + dataToSend
    sacp := `{"cmd":"sacp",` + dataToMe
    facp := `{"cmd":"facp",` + dataToMe
    
    res := TransmitData(ip, MAINLISTEN, racp)
    if res {
        TransmitData(myIP, SRLISTEN, sacp)
        TransmitData(ip, SRLISTEN, racp)
    }else{
        TransmitData(ip, SRLISTEN, facp)
        TransmitData(myIP, SRLISTEN, facp)
    }
}

func SendReject(ip, mac, dir, username string){
    fileName := utils.GetFileName(dir)
    fileType := utils.GetFileExt(fileName)

    dataToSend := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","contentType":"file"}`,
     myUsername, myIP, myMAC, dir, fileName, fileType)
    dataToMe := fmt.Sprintf(`"username":"%v","ip":"%v","mac":"%v","dir":"%v","fileName":"%v","fileType":"%v","contentType":"file"}`,
     username, ip, mac, dir, fileName, fileType)

    rrej := `{"cmd":"rrej",` + dataToSend
    srej := `{"cmd":"srej",` + dataToMe
    frej := `{"cmd":"frej",` + dataToMe

    res := TransmitData(ip, SRLISTEN, rrej)
    if res {
    	TransmitData(myIP, SRLISTEN, srej)
    }else{
    	TransmitData(myIP, SRLISTEN, frej)
    }
}

func SendMessage(ip, mac, username, msg string){
    dataToSend := fmt.Sprintf(`"mac":"%v","person":"%v","content":"%v","contentType":"text"}`, myMAC, myUsername, msg)
    dataToMe := fmt.Sprintf(`"mac":"%v","person":"%v","content":"%v","contentType":"text"}`,mac,  username, msg)

    rmsg := `{"cmd":"rmsg",` + dataToSend
    smsg := `{"cmd":"smsg",` + dataToMe
    fmsg := `{"cmd":"fmsg",` + dataToMe

    res := TransmitData(ip,MSGLISTEN, rmsg)
    if res {
        TransmitData(myIP, MSGLISTEN, smsg)
    }else{
        TransmitData(myIP, MSGLISTEN, fmsg)
    }
}

func TransmitData(ip string, port int, msg string) bool{
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