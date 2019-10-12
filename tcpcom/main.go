package tcpcom


import (
	"fmt"
	"net"
    "io/ioutil"
    "strconv"
    "time"
	utils "wsftp/utils"
)


// ports
var mainComand int = 9999
// 10000 reserved for handshake protocol
// SendReceive port
var srListen int = 10001
// Messaging port
var msgListen int = 10002

// others
var MB int = 1048576 // 1MB
var SpeedControlPortLimit int64 = 10485760 // 10MB
var StandartSpeed int64 = 3670016 // 3.5MB
// speed test count for average
var NTest int = 5
// my local ip
var myIP string = utils.GetInterfaceIP().String()
// computer username
var username string = utils.GetUsername()
// fail message stores last message for safely sending last info of the program
var failMsg string = "initial fail message"


type comm struct {
	ip string
	port int
	conn *net.Conn
} 

// constructor of this package
func NewComm(IP string, port int) *comm{
	pc := comm{ip:IP, port:port}
	return &pc
}

// Sending Header information function
func (c *comm) Header(dir , dest string) bool{
	data := PackHeader(dir, dest)
    res := send(c.ip, c.port, data)
    return res
}

// sending acknowledge
func (c *comm) Ack() bool{
	data := []byte{1}
	res := send(c.ip, c.port, data)
	return res
}

// sending not acknowledge
func (c *comm) Nack() bool{
	data := []byte{0}
	res := send(c.ip, c.port, data)
	return res
}

// speed test core function
func (c *comm)sendTest(ch chan<- float64) bool{
	data := make([]byte, MB)
	start := time.Now()
    res := send(c.ip, c.port, data)
	end := time.Now()
    ch <- float64(1 / end.Sub(start).Seconds())
    return res
}

// sending int64
func (c *comm) SendInt(offset int64) bool{
	data := utils.IntToByteArray(offset, 8)
    res := send(c.ip, c.port, data)
    return res
}

// speedtest main function
func (c *comm) SpeedTest(ch chan<- int64) bool{
	sumMbs := float64(0)
	fl64 := make(chan float64, 1)
	bch := make(chan bool, 1)
	for i := 0; i < NTest ; i++ {
		res := c.sendTest(fl64)
		if res {
			sumMbs += <- fl64
		}else{
			return false
		}
		res2 := c.Rec(bch)
		<- bch
		if !res2 {
			return false
		}

	}
	ch <- int64(sumMbs / float64(NTest) * float64(MB))
	return true
}

// Receive header function
func (c *comm) RecHeader() (string, string, string, int64, bool){
	ch := make(chan []byte, 1)
	res := receive(c.port, ch)
	msg := <- ch
	n, n1, n2, n3 := UnpackHeader(msg)
	return n, n1, n2, n3, res
}

// Receive test function
func (c *comm) recTest() bool{
	ch := make(chan []byte, 1)
	res := receive(c.port, ch)
	<-ch
	return res
}

// Receive whole speed test
func (c *comm) RecSpeedTest() bool{
	for i := 0 ; i < NTest ; i++ {
		res := c.recTest()
		if !res {
			return false
		}
		res2 := c.Ack()
		if !res2{
			return false
		}
	}
	return true
}

// Receive bool , for ack and nack
func (c *comm) Rec(bch chan<- bool) bool{
	ch := make(chan []byte, 1)
	res := receive(c.port, ch)
	msg := <- ch
	if msg[0] == byte(5) && len(msg) == 1{
		bch <- true
	}else{
		bch <- false
	}
	return res
}

// receive in64
func (c *comm) RecInt(ch64 chan<- int64) bool{
	ch := make(chan []byte, 1)
	res := receive(c.port, ch)
	msg := <- ch
	ch64 <- utils.ByteArrayToInt(msg)
	return res
}

// receive byte array
func (c *comm) RecData(chbyte chan<- []byte) bool{
	res := receive(c.port, chbyte)
	return res
}

// send byte array
func (c * comm) SendData(data []byte) bool{
    res := send(c.ip, c.port, data)
    return res
}

func ReceiveFile(ip string, port int, controlPort * int){
	boolChan := make(chan bool, 1)
	byteChan := make(chan []byte, 1)
	ch64 := make(chan int64, 1)

	// fail message for backup
	failMsg = fmt.Sprintf(`{"stat":"ncon","ip":"%v","port":"%v","info":"receive start"}`, ip, port) 

	com := NewComm(ip, port)

	// send header
	otheruser , dest, fileName, fileSize, res := com.RecHeader()
	if !res {return}

	dir := dest + utils.Sep + fileName

	offset := utils.GetFileSize(dir)

	remaining := fileSize - offset

	if remaining == 0 && utils.IsFileExist(dir){
		fileName = utils.UniqName(dest, fileName)
		dir = dest + utils.Sep + fileName
		remaining = fileSize
		offset = 0
	}

	// fail message for backup
	failMsg = fmt.Sprintf(`{"stat":"fprg",username":"%v","ip":"%v","port":"%v","dir":"%v",total":"%v","current":"%v","speed":"%v"}`, username,  myIP, port, dir, 0, 0, 0)

	// send offset
	res = com.SendInt(offset)
	if !res {return}

	// speed test
	if remaining >= SpeedControlPortLimit {
		res = com.RecSpeedTest()
		if !res {return}
	}

	// receive
	res = com.Rec(boolChan)
	if res && <-boolChan{return}

	// ack
	res = com.Ack()
	if !res {return}

	// receive int64
	res = com.RecInt(ch64)
	n := <- ch64
	if !res {return}

	// ack
	res = com.Ack()
	if !res {return}

	comSR := NewComm(ip,srListen)
	comMYSR := NewComm(myIP,srListen)

	done := false

	for i := int64(0) ; i < n ; i++ {
		if *controlPort == 0{
			return
		}else{
			// receive data
			res = com.RecData(byteChan)
			if res {
				data := <- byteChan
				utils.FWrite(dir, data)
			}else{
				return
			}
			// ack
			res = com.Ack()
			if !res {return}
		}
		if i == n - 1{
			done = true
		}
	}
	mymsg := fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v"}`, otheruser, com.ip, com.port, dir)
	othermsg := fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v"}`, username, myIP, com.port, dir)

	if done == false && *controlPort == 0 {
		res = comMYSR.SendData([]byte(`{"stat":"kprg",` + mymsg))
		if !res {return}
		res = comSR.SendData([]byte(`{"stat":"kprg",` + othermsg))
		if !res {return}
	}
}

func SendFile(ip string, port int, dir string, dest string, controlPort * int){

	// fail message for backup
	failMsg = fmt.Sprintf(`{"stat":"ncon","ip":"%v","port":"%v","info":"send start"}`, myIP, port) 

	boolChan := make(chan bool, 1)
	ch64 := make(chan int64, 1)

	com := NewComm(ip, port)

	// sending header
	res := com.Header(dir, dest)
	if !res {return}

	
	// receiveing offset
	res = com.RecInt(ch64)
	off := <- ch64
	if !res {return}
	
	// speed test controlPort mechanism
	fileSize := utils.GetFileSize(dir)
	remaining := fileSize - off
	speed := int64(0)
	if remaining >= SpeedControlPortLimit {
		res = com.SpeedTest(ch64)
		if !res {return}else{speed = <- ch64}
	}else{
		speed = StandartSpeed
	}
	
	res = com.Ack()
	if !res {return}
	
	res = com.Rec(boolChan)
	if res && <-boolChan{return}

	n := utils.GetPackNumber(remaining, speed)
		// total pack number

	// send pack size
	res = com.SendInt(int64(n))
	if !res {return}
	// com.SendInt(int64(n))

	// ack for receiving n
	res = com.Rec(boolChan)
	if res && <-boolChan{return}

	comSR := NewComm(ip,srListen)
	comMYSR := NewComm(myIP,srListen)
	comMYCMD := NewComm(myIP,mainComand)


	total := fileSize
	data := make([]byte, 0, speed)
	for i := 0 ; i < n ; i ++ {
		if *controlPort == 0 {
			otherSR := fmt.Sprintf(`username":"%v","ip":"%v","port":"%v","dir":"%v"}`, username, myIP, com.port, dir)
			res = comSR.SendData([]byte(`{"stat":"kprg",` + otherSR))
			if !res {return}
			mySR := fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v"}`, username, com.ip, com.port, dir)
			res = comMYSR.SendData([]byte(`{"stat":"kprg",` + mySR))
			if !res {return}
			return
		}else{
			if i == n - 1 {
				data = utils.FReadAt(dir, off, fileSize - off)
				off += fileSize - off
			}else{
				data = utils.FReadAt(dir, off, speed)
				off += speed
			}
			start := time.Now()
			res = com.SendData(data)
			if !res {return}
			end := time.Now()
			res = com.Rec(boolChan)
			if res && <-boolChan{return}
			delt := int64(end.Sub(start))
			instatSpeed := float64(delt) / 1e9 * float64(speed) / float64(MB) * 1024
			otherSR := fmt.Sprintf(`username":"%v","ip":"%v","port":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v"}`, username, myIP, com.port, dir, total, off, int(instatSpeed))
			res = comSR.SendData([]byte(`{"stat":"rprg",` + otherSR))
			if !res {return}
			mySR := fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v"}`, username, com.ip, com.port, dir, total, off, int(instatSpeed))
			res = comMYSR.SendData([]byte(`{"stat":"sprg",` + mySR))
			if !res {return}
			// fail message for backup
			failMsg = `{"stat":"fprg",` + mySR
		}
	}
	msg := fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v"}`, username, myIP, com.port, dir, total, total, 0)
	res = comSR.SendData([]byte(`{"stat":"dprg",` + msg))
	if !res {return}

	msg = fmt.Sprintf(`"username":"%v",ip":"%v","port":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v"}`, username, com.ip, com.port, dir, total, total, 0)
	res = comMYSR.SendData([]byte(`{"stat":"dprg",` + msg))
	if !res {return}

	res = comMYCMD.SendData([]byte(`{"stat":"dprg",` + msg))
	if !res {return}
}

func send(ip string, port int, data []byte) bool{
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
		send(myIP, mainComand, []byte(failMsg))
		// send(ip, mainComand, []byte(failMsg))
		return false
    }else{
        conn.Write(data)
        conn.Close()
        return true
    }
}

func receive(port int, ch chan<- []byte) bool {
    strPort := strconv.Itoa(port)
    addr := myIP + ":" + strPort
    tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
        fmt.Println("Address resolving error (Inner)",err)
		ch <- []byte{0}
        return false
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        fmt.Println("Listen Error (Inner)", err)
		send(myIP, mainComand, []byte(failMsg))
		listener.Close()
		ch <- []byte{0}
		return false

    }else{
    	listener.SetDeadline(time.Now().Add(10 * time.Second))
        defer listener.Close()
    }
    conn, err := listener.Accept()
    if err != nil {
        fmt.Println("Listen Accept Error (Inner) ", failMsg)
		send(myIP, mainComand, []byte(failMsg))
		return false
    }
    msg, err :=  ioutil.ReadAll(conn)
    if err != nil {
        fmt.Println("Message Read Error (Inner)", err)
    }
    conn.Close()
    ch <- msg
	return true
}