package hs

import (
	"fmt"
	"net"
	"net/http"
	"os"
    "os/signal"
    "syscall"
    "wsftp/utils"
	"github.com/gorilla/websocket"
	"github.com/ecoshub/jparse"
)

const (
	MAINPORT string = "9998"
	WSCOMPORT string = "10000"
	ENDPOINT = "/hs"
	BROADCASTLISTENIP string = "0.0.0.0"
	LOOPCONTROLLIMIT int = 100
	UDPREPEAT int = 5
)

var (
	broadcastIP string = utils.GetBroadcastIP().String()
	myIP string = utils.GetInterfaceIP().String()
	myEthMac string = utils.GetEthMac()
	receiveControl bool = true
	MACList []string = make([]string,0,1024)
	onlineCount int = 0
	myUsername string = utils.GetCustomUsername()
	onlines = make(map[string][]string, 256)
	innerMessageChan = make(chan []byte, 1)
	sigs = make(chan os.Signal, 1)

	upgraderHS = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: false,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	}

)

func handleConn(w http.ResponseWriter, r *http.Request){
	ws, err := upgraderHS.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Handshake web socket connection error: ", err)
	}
	defer ws.Close()
	for {
		ws.WriteMessage(1, []byte(<-innerMessageChan))
	}
}

func Start(){
	http.HandleFunc(ENDPOINT, handleConn)
	go activity()
	err := http.ListenAndServe(":" + WSCOMPORT, nil)
	fmt.Println("Handshake server shutdown unexpectedly!", err)
}

func Restart(){
    done := make(chan bool, 1)
    onClose(done)
    <-done
    myUsername = utils.GetCustomUsername()
	MACList = make([]string,0,1024)
    onlines = make(map[string][]string, 128)
    data := fmt.Sprintf(`{"event":"online","ip":"%v","username":"%v","mac":"%v"}`, myIP, myUsername, myEthMac)
    sendPack(broadcastIP, MAINPORT, []byte(data))
}

func activity(){	
    done := make(chan bool, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    data := fmt.Sprintf(`{"event":"online","ip":"%v","username":"%v","mac":"%v"}`, myIP, myUsername, myEthMac)
    sendPack(broadcastIP, MAINPORT, []byte(data))
	receiveChan := make(chan []byte, 1)
    go func() {
        <- sigs
        onClose(done)
        <- done
      	os.Exit(0)
    }()
	for receiveControl {
		go receive(BROADCASTLISTENIP, MAINPORT, receiveChan)
		receive := <- receiveChan
		if receive[0] != 0 {
			data := <- receiveChan
			json := jparse.Parse(data)
			tempStatus, _ := json.GetString("event")
			tempIP, _ := json.GetString("ip")
			tempUsername, _ := json.GetString("username")
			tempMAC, _ := json.GetString("mac")
			msg := fmt.Sprintf(`{"event":"%v","ip":"%v","username":"%v","mac":"%v"}`,tempStatus, tempIP, tempUsername, tempMAC)
			if tempMAC != myEthMac {
				if !hasThis(MACList, tempMAC) && tempStatus == "online"{
					onlines[tempMAC] = []string{tempUsername, tempIP} 
					MACList = append(MACList, tempMAC)
					onlineCount++
					innerMessageChan <- []byte(msg)
					data := fmt.Sprintf(`{"event":"online","ip":"%v","username":"%v","mac":"%v"}`, myIP, myUsername, myEthMac)
		    		sendPack(broadcastIP, MAINPORT, []byte(data))
				}
				if hasThis(MACList, tempMAC) && tempStatus == "offline"{
					MACList = removeFromList(MACList, tempMAC)
					delete(onlines, tempMAC) 
					onlineCount--
					innerMessageChan <- []byte(msg)
				}
			}
		}
	}
}

func GetIP(mac string) string{
	if len(onlines[mac]) != 0 {
		return onlines[mac][1]
	}else{
		sendInfo("No IP address match!")
	}
	return ""
}

func GetUsername(mac string) string{
	if len(onlines[mac]) != 0 {
		return onlines[mac][0]
	}else{
		sendInfo("No MAC address match!")
	}
	return ""
}

func receive(ip, port string, ch chan<- []byte){
	buff := make([]byte, 1024)
    pack, err := net.ListenPacket("udp", ip + ":" + port)
    if err != nil {
        sendInfo("UDP(R) Connection Error " + err.Error())
        pack.Close()
        ch <- []byte{0}
    }
    n, _, err := pack.ReadFrom(buff)
    if err != nil {
        sendInfo("UDP(R) Read Error " + err.Error())
        pack.Close()
        ch <- []byte{0}
    }
	defer pack.Close()
    ch <- buff[:n]
}

func sendPack(ip, port string, data []byte){
	sendValidationChan := make(chan int, 1)
	valid := 0
	count := 0
	for valid != 1 {
		for i := 0 ; i < UDPREPEAT ; i++ {
			go send(ip, port, data , sendValidationChan)
		}
		valid = <- sendValidationChan
		count++
		if count > LOOPCONTROLLIMIT {
        	sendInfo("UDP(S) Repetition Error. Something is wrong can't send any signal!")
			return
		}
	}
}

func send(ip, port string, data []byte, ch chan<- int){
    conn, err := net.Dial("udp", ip + ":" + port)
       if err != nil {
        sendInfo("UDP(S) Connection Error." + err.Error())
    	ch <- 0
    }else{
    	defer conn.Close()
	    conn.Write(data)
	    ch <- 1
    }
}

func onClose(ch chan<- bool){
	sendValidationChan := make(chan int, 1)
	data := fmt.Sprintf(`{"event":"offline","ip":"%v","username":"%v","mac":"%v"}`, myIP, myUsername, myEthMac)
	valid := 0
	count := 0
	for valid != 1 {
		for i := 0 ; i < UDPREPEAT ; i++ {
			go offlineFunc(broadcastIP, MAINPORT, []byte(data) , sendValidationChan)
		}
		valid = <- sendValidationChan
		count++
		if count > LOOPCONTROLLIMIT {
        	sendInfo("UDP(S-Off) Repetition Error. Something is wrong can't send any signal!")
			break
		}
	}
	ch <- true
}

func offlineFunc(ip, port string, data []byte, ch chan<- int){
    conn, err := net.Dial("udp", ip + ":" + port)
       if err != nil {
        sendInfo("UDP(S-Off) Connection Error." + err.Error())
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

func sendInfo(msg string){
    msg = fmt.Sprintf(`{"event":"info","content":"-HANDSHAKE- %v"}`, msg)
	innerMessageChan <- []byte(msg)
}