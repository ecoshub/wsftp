package hs

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"os"
    "os/signal"
    "syscall"
    "wsftp/utils"
	"github.com/gorilla/websocket"
)

const (
	port string = "9998"
	hsPort string = "10000"
	endPoint = "/hs"
	broadcastListenIP string = "0.0.0.0"
	loopControl int = 100
	udpRepeat int = 5
)

var (
	broadcastIP string = utils.GetBroadcastIP().String()
	myIP string = utils.GetInterfaceIP().String()
	myIPB []byte = []byte(myIP)
	myEthMac string = utils.GetEthMac()
	myEthMacB []byte = []byte(myEthMac)
	receiveControl bool = true
	IPList []string = make([]string,0,1024)
	onlineCount int = 0
	UsernameList []string = make([]string,0,1024)
	myUsername string = utils.GetUsername()
	myUsernameB []byte = []byte(myUsername)
	msgOn []byte = []byte("online")
	msgOff []byte = []byte("offline")

	messageChan = make(chan []byte, 1)

	upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}
)


func handleConn(w http.ResponseWriter, r *http.Request){
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Handshake web socket connection error: ", err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-messageChan))
	}
}

func Start(){
	http.HandleFunc(endPoint, handleConn)
	go activity()
	err := http.ListenAndServe(":" + hsPort, nil)
	fmt.Println("Handshake server shutdown unexpectedly!", err)
}

func activity(){	
	sigs := make(chan os.Signal, 1)
    done := make(chan bool, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	data := concatByteArray(" ", msgOn, myUsernameB, myIPB, myEthMacB)
    sendPack(broadcastIP, port, data)
	receiveChan := make(chan string, 1)
    go func() {
        <- sigs
        onClose(done)
        <- done
      	os.Exit(0)
    }()
	for receiveControl {
		go receive(broadcastListenIP, port, receiveChan)
		tempPack := <- receiveChan
		tempMsg, tempIP, tempUsername, tempMAC := parsePack(tempPack)
		msg := fmt.Sprintf(`{"stat":"%v","ip":"%v","username":"%v","mac":"%v"}`,tempMsg, tempIP, tempUsername, tempMAC)
		if !hasThis(IPList, tempIP) && tempIP != myIP && tempMsg == string(msgOn){
			IPList = append(IPList, tempIP)
			UsernameList = append(UsernameList, tempUsername)
			onlineCount++
			messageChan <- []byte(msg)
			data := concatByteArray(" ", msgOn, myUsernameB, myIPB, myEthMacB)
    		sendPack(broadcastIP, port, data)
		}
		if hasThis(IPList, tempIP) && tempIP != myIP && tempMsg == string(msgOff){
			IPList = removeFromList(IPList, tempIP)
			UsernameList = removeFromList(UsernameList, tempUsername)
			messageChan <- []byte(msg)
		}
	}
}


func receive(ip, port string, ch chan<- string){
	buff := make([]byte, 1024)
    pack, err := net.ListenPacket("udp", ip + ":" + port)
    if err != nil {
        fmt.Println("Connection Fail", err)
    }
    n, addr, err := pack.ReadFrom(buff)
    if err != nil {
        fmt.Println("Read Error", err)
    }else{
    	defer pack.Close()
	    ipandport := strings.Split(addr.String(), ":")
	    remoteIP := ipandport[0]
	    buff = buff[:n]
	    ch <- string(buff) + " " + remoteIP
    }
}

func sendPack(ip, port string, data []byte){
	sendValidationChan := make(chan int, 1)
	valid := 0
	count := 0
	for valid != 1 {
		for i := 0 ; i < udpRepeat ; i++ {
			go send(broadcastIP, port, data , sendValidationChan)
		}
		valid = <- sendValidationChan
		count++
		if count > loopControl {
			fmt.Println("Something is wrong can't send any signal!")
			return
		}
	}

}

func send(ip, port string, data []byte, ch chan<- int){
    conn, err := net.Dial("udp", ip + ":" + port)
       if err != nil {
        fmt.Println("Connection Fail")
    	ch <- 0
    }else{
    	defer conn.Close()
	    conn.Write(data)
	    ch <- 1
    }
}

func onClose(ch chan<- bool){
	sendValidationChan := make(chan int, 1)
	data := concatByteArray(" ", msgOff, myUsernameB)
	valid := 0
	count := 0
	for valid != 1 {
		for i := 0 ; i < udpRepeat ; i++ {
			go offlineFunc(broadcastIP, port, data , sendValidationChan)
		}
		valid = <- sendValidationChan
		count++
		if count > loopControl {
			fmt.Println("Something is wrong can't send any signal!")
			break
		}
	}
	ch <- true
}


func offlineFunc(ip, port string, data []byte, ch chan<- int){
    conn, err := net.Dial("udp", ip + ":" + port)
       if err != nil {
        fmt.Println("Connection Fail")
        ch <- 0
    }else{
    	defer conn.Close()
    	conn.Write(data)
    	ch <- 1
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

func parsePack(pack string) (msg string, IP string, username string, MAC string){
	tokens := strings.Split(pack, " ")
	msg = tokens[0]
	username = tokens[1]
	IP = tokens[2]
	MAC = tokens[3]
	return
}

func removeFromList(list []string, el string) []string{
	lenl := len(list)
	if lenl < 2 {
		return nil
	}
	newList := make([]string,lenl - 1,1024)
	count := 0
	for _, v := range list {
		if v != el  {
			newList[count] = v
			count++
		}
	}
	return newList
}

func concatByteArray(sep string, arr ...[]byte) []byte {
	newArr := make([]byte ,0 ,1024)
	lena := len(arr)
	sepB := []byte(sep)
	for i, v := range arr {
		newArr = append(newArr, v...)
		if i != lena - 1 {
			newArr = append(newArr, sepB...)
		} 
	}
	fmt.Println(newArr)
	fmt.Println(string(newArr))
	return newArr
}