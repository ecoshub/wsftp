package tcpcom

import (
	"fmt"
	"net"
    "strconv"
    "time"
	utils "wsftp/utils"
	rw "github.com/ecoshub/penman"
)

const (
	// ports
	MAINCOMANDPORT int = 9999
	// 10000 reserved for handshake protocol
	SRLISTENPORT int = 10001
	MSGLISTENPORT int = 10002

	// Data Size
	MB int = 1048576 // 1MB
	SPEEDTESTLIMIT int = MB * 1000 // 1000MB for VM
	// SPEEDTESTLIMIT int = MB * 10 // 10MB
	STANDARTSPEED int = MB * 5 // 4MB
	TCPREADSIZE int = 65536 // as byte
	// write RAM to DISK treshold
	WRITEDISCBUFFER int = MB * 10 // as byte
	READDISCBUFFER int = MB * 10 // as byte

	// general settings
	WRITEREPETITION int = 3
	WRITEREPETITIONDELAY int = 100 // as ms
	TCPDEADLINE int = 5 // as second
)

var (
	// enviroment
	myIP string = utils.GetInterfaceIP().String()
	myMac string = utils.GetEthMac()

	// for debug
	step string = "::Steps Start::\n"
	stepCount int = 0
)

// master comminication struct
type comm struct {
	ip string
	port int
	conn net.Conn
} 

func NewCom(ip string, port int) *comm{
	tempCon := comm{ip:ip,port:port}
	return &tempCon
}

func SendFile(ip, mac, username string, port int, id, dir, dest string, control * int){

	boolChan := make(chan bool, 1)
	intChan := make(chan int, 1)
	int64Chan := make(chan int64, 1)

	// main comminication struct
	com := NewCom(ip, port)
	
	fileSize := utils.GetFileSize(dir)
	filename := utils.GetFileName(dir)
	
	// dial to receiver
	res := com.Dial()
	if !res {return}

	// send dest
	res = com.SendData([]byte(dest))
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filename
	res = com.SendData([]byte(filename))
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filesize
	res = com.SendInt(fileSize)
	if !res {return}

	// receive offset
	res = com.RecInt(int64Chan)
	if !res {return}
	off := <-int64Chan

	// speed test controlPort mechanism
package tcpcom

import (
	"fmt"
	"net"
    "strconv"
    "time"
	utils "wsftp/utils"
	rw "github.com/ecoshub/penman"
)

const (
	// ports
	MAINCOMANDPORT int = 9999
	// 10000 reserved for handshake protocol
	SRLISTENPORT int = 10001
	MSGLISTENPORT int = 10002

	// Data Size
	MB int = 1048576 // 1MB
	SPEEDTESTLIMIT int = MB * 1000 // 1000MB for VM
	// SPEEDTESTLIMIT int = MB * 10 // 10MB
	STANDARTSPEED int = MB * 5 // 4MB
	TCPREADSIZE int = 65536 // as byte
	// write RAM to DISK treshold
	WRITEDISCBUFFER int = MB * 10 // as byte
	READDISCBUFFER int = MB * 10 // as byte

	// general settings
	WRITEREPETITION int = 3
	WRITEREPETITIONDELAY int = 100 // as ms
	TCPDEADLINE int = 5 // as second
)

var (
	// enviroment
	myIP string = utils.GetInterfaceIP().String()
	myMac string = utils.GetEthMac()

	// for debug
	step string = "::Steps Start::\n"
	stepCount int = 0
)

// master comminication struct
type comm struct {
	ip string
	port int
	conn net.Conn
} 

func NewCom(ip string, port int) *comm{
	tempCon := comm{ip:ip,port:port}
	return &tempCon
}

func SendFile(ip, mac, username string, port int, id, dir, dest string, control * int){

	boolChan := make(chan bool, 1)
	intChan := make(chan int, 1)
	// int64Chan := make(chan int64, 1)

	// main comminication struct
	com := NewCom(ip, port)
	
	fileSize := utils.GetFileSize(dir)
	filename := utils.GetFileName(dir)
	
	// dial to receiver
	res := com.Dial()
	if !res {return}

	// send dest
	res = com.SendData([]byte(dest))
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filename
	res = com.SendData([]byte(filename))
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filesize
	res = com.SendInt(fileSize)
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// speed test controlPort mechanism
	// fileSize = utils.GetFileSize(dir)
	// remaining := fileSize - off

	speed := int64(0)
	if int(fileSize) >= SPEEDTESTLIMIT {
		// run speed test
		res = com.SendTestData(intChan)
		if !res {return}else{speed = int64(<- intChan)}

	}else{
		// ack
		res = com.Ack()
		if !res {return}
		speed = int64(STANDARTSPEED)
	}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filesize
	res = com.SendInt(speed)
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	batchSize := utils.GetPackNumber(fileSize, int64(READDISCBUFFER))
	innerBatchSize := utils.GetPackNumber(int64(READDISCBUFFER), speed)

	// // total := fileSize
	data := make([]byte, speed)
	datalen := int64(READDISCBUFFER)

	off := 0
	for i := 0 ; i < batchSize ; i++ {
		if *control == 1{
			if i == batchSize - 1 {
				data = rw.ReadAt(dir, off, fileSize - off)
				datalen = fileSize - off
				innerBatchSize = utils.GetPackNumber(datalen, speed)
			}else{
				data = rw.ReadAt(dir, off, int64(READDISCBUFFER))
			}
			innerData := make([]byte, 0, speed)
			for j := 0 ; j < innerBatchSize ; j++ {
				if datalen > speed {
					if j == innerBatchSize - 1{
						innerData = data[int(speed) * (j):]
						off += int64(len(innerData))
					}else{
						innerData = data[int(speed) * (j):int(speed) * (j+ 1)]
						off += int64(speed)
					}
				}else{
					innerData = data
					off += int64(len(innerData))
				}
				start := time.Now()
				res = com.SendData(innerData)
				end := time.Now()
				if !res {*control = 0;break}

				currentSpeed := float64(speed) / float64(end.Sub(start).Seconds() * 1e3) // kb/second 

				msg := fmt.Sprintf(`{"event":"prg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
				 username, ip, mac, port, id, dir, fileSize, off, int(currentSpeed))

				SendMsg(myIP, SRLISTENPORT, msg)
			}
		}else{

			msg := fmt.Sprintf(`{"event":"fprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
			 username, ip, mac, port, id, dir, fileSize, off, 0)
			SendMsg(myIP, MAINCOMANDPORT, msg)
			time.Sleep(10 * time.Millisecond)
			SendMsg(myIP, SRLISTENPORT, msg)
			com.Close()
			return
		}
	}
	msg := fmt.Sprintf(`{"event":"dprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
	 username, ip, mac, port, id, dir, fileSize, off, 0)
	SendMsg(myIP, MAINCOMANDPORT, msg)
	time.Sleep(10 * time.Millisecond)
	SendMsg(myIP, SRLISTENPORT, msg)
	com.Close()
}


func ReceiveFile(ip, mac, username string, port int, id string, control * int){

	byteChan := make(chan []byte, 1)
	boolChan := make(chan bool, 1)
	int64Chan := make(chan int64, 1)

	// main comminication struct
	com := NewCom(ip, port)

	// listen start
	res := com.Listen()
    if !res{return}

    // receive file destination
    res = com.RecData(byteChan)
    if !res {return}

    dest := string(<- byteChan)

    // ack
    res = com.Ack()
    if !res{return}

    // receive filename
    res = com.RecData(byteChan)
    if !res {return}

    filename := string(<- byteChan)

    // ack
    res = com.Ack()
    if !res{return}

    // receive file size
    res = com.RecInt(int64Chan)
    if !res {return}
	
	fileSize := <-int64Chan

    filename = utils.UniqName(dest, filename, fileSize)
    dir := dest + utils.Sep + filename
	// offset := utils.GetFileSize(dir)
	// remaining := fileSize - offset

    // ack
    res = com.Ack()
    if !res{return}

    // send offset of file
    // res = com.SendInt(0)
    // if !res {return}

    // if filesize bigger than speed test limit run a speed test
    if int(fileSize) >= SPEEDTESTLIMIT {
        res = com.RecTestData()
        if !res {return}
    }else{
        res = com.Rec(boolChan)
        if !res{return}else{<-boolChan}
    }

    // ack
    res = com.Ack()
    if !res{return}

    // receive speed
    res = com.RecInt(int64Chan)
    if !res {return}
	
	speed := <-int64Chan

    // ack
    res = com.Ack()
    if !res{return}	

   // download file
    count := fileSize
    currentSize := int64(0)
    mainBuffer := make([]byte, 0, WRITEDISCBUFFER)
    genCount := 0
    currentSpeed := float64(0)
	printCount := int64(0)
    start := time.Now()
    for count > 0 {
    	if *control == 1 {
	        n, res := com.Read(byteChan, "Receive File Data")
	        if res {
	            mainBuffer = append(mainBuffer, (<-byteChan)...)
	        }else{*control = 0}
		
	        if genCount > WRITEDISCBUFFER{
	            rw.Write(dir, mainBuffer)
	            mainBuffer = make([]byte, 0, WRITEDISCBUFFER )
	            genCount = 0
	        }
			if printCount > speed{
		    	end := time.Now()
				currentSpeed = float64(printCount) / float64(end.Sub(start).Seconds() * 1e3)
				msg := fmt.Sprintf(`{"event":"prg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
				 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
				SendMsg(myIP, SRLISTENPORT, msg)
		    	start = end
				printCount = 0
			}			
	        count -= int64(n)
	        genCount += n
			printCount += int64(n)
    		currentSize = fileSize - count
    	}else{
			msg := fmt.Sprintf(`{"event":"fprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
			 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
			SendMsg(myIP, MAINCOMANDPORT, msg)
			time.Sleep(10 * time.Millisecond)
			SendMsg(myIP, SRLISTENPORT, msg)
			com.Close()
			done := rw.DelFile(dir)
			if done {
				SendMsg(myIP, SRLISTENPORT, fmt.Sprintf(`{"event":"info","content":"Unfinished file deleted. directory:%v"}`, dir))
			}else{
				SendMsg(myIP, SRLISTENPORT, fmt.Sprintf(`{"event":"info","content":"Unfinished file delete operation fail. directory:%v"}`, dir))
			}
			return
    	}
    }
    if len(mainBuffer) > 0 {
        rw.Write(dir, mainBuffer)
    }
	msg := fmt.Sprintf(`{"event":"dprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
	 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
	SendMsg(myIP, MAINCOMANDPORT, msg)
	time.Sleep(10 * time.Millisecond)
	SendMsg(myIP, SRLISTENPORT, msg)
    com.Close()
}
// Receive bool , for ack and nack functions
func (c * comm) Rec(ch chan<- bool) bool{
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan, "Receive Ack/Nack")
	if !res {return false}
	byt := (<- byteChan)[0]
	if byt == 1 {
		ch <- true
	}else{
		ch <- false
	}
	return true
}

// send acknowledge
func (c * comm) Ack() bool{
	return c.Write([]byte{1}, "Send Ack")
}

// send not acknowledge
func (c * comm) Nack() bool{
	return c.Write([]byte{0}, "Send Nack")
}

// receive int64
func (c *comm) RecInt(int64Chan chan<- int64) bool{
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan, "Receive INT64")
	if res {
		numInt, _ := strconv.Atoi(string(<-byteChan))
		num := int64(numInt)
		int64Chan <- num
		return true
	}else{
		int64Chan <- int64(-1)
	}
	return false
}

// send int64
func (c *comm) SendInt(number int64) bool{
	strNum := strconv.FormatInt(number, 10)
	data := []byte(strNum)
    res := c.Write(data, "Send INT64")
    return res
}

// receive byte array
func (c *comm) RecData(chbyte chan<- []byte) bool{
	_, res := c.Read(chbyte, "Receive Data")
	return res
}

// send byte array
func (c * comm) SendData(data []byte) bool{
    return c.Write(data, "Send Data")
}

// Receive speed test data
func (c *comm) RecTestData() bool{
	byteChan := make(chan []byte, 1)
	count := MB
	for count > 0 {
		n, res := c.Read(byteChan, "Receive Test Data")
		if !res{return false}
		count -= n
		<-byteChan
	}
	return true
}

// Send speed test data
func (c *comm) SendTestData(ch chan<- int) bool{
	data := make([]byte, MB)
	res := false
	start := time.Now()
	res = c.SendData(data)
	if !res {return false}
	end := time.Now()
	ch <- int(float64(MB) / end.Sub(start).Seconds())
	return true
}

func (c * comm) Dial() bool{
	strPort := strconv.Itoa(c.port)
	conn, err := net.Dial("tcp", c.ip + ":" + strPort)
	if err != nil {
		fmt.Printf("Main Fail:\"Dial Connection\", Error String: %v\n", err.Error())
		return false
	}else{
		conn.SetReadDeadline(time.Now().Add(time.Duration(TCPDEADLINE) * time.Second))
	    c.conn = conn
	    return true
	}
}

func (c * comm) Listen() bool{
    strPort := strconv.Itoa(c.port)
    addr := myIP + ":" + strPort
    tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
		fmt.Printf("Main Fail:\"Adress Resolving\", Error String: %v\n", err.Error())
		return false
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
		fmt.Printf("Main Fail:\"Listen Error\", Error String: %v\n", err.Error())
		return false
    }else{
    	listener.SetDeadline(time.Now().Add(time.Duration(TCPDEADLINE) * time.Second))
	    conn, err := listener.Accept()
	    if err != nil {
			fmt.Printf("Main Fail:\"Listen Accept Error\", Error String: %v\n", err.Error())
			return false
	    }else{
	    	c.conn = conn
	    	return true
	    }
    }
}

func (c * comm) Write(msg []byte, label string) bool{
	// for debug
	stepCount++
	step += strconv.Itoa(stepCount) + " >> " + label + "\n"
	innerlabel := ""
	// for debug

	// FOR VM
	time.Sleep(100 * time.Millisecond)
	res := false
	for i := 0 ; i < WRITEREPETITION ; i ++ {
		// mini delay for try again
		time.Sleep(time.Duration(WRITEREPETITIONDELAY) * time.Millisecond)
		res, innerlabel = c.writeCore(msg)
		if res {
			return true
		}
	}
	// for debug
	fmt.Println(step)
	// for debug
	fmt.Printf("Main Fail:\"%v\", Inner Fail:\"%v\", Write Repetition:%v\n", innerlabel, label, WRITEREPETITION)
	return false
}

func (c * comm) writeCore(data []byte) (bool, string){
    label := ""
	n, err := c.conn.Write(data)
	if err != nil {
		// for debug
    	label = "Dial Connection " + err.Error()
    	return false, label
	}else{
		if n != len(data){
			return false, label
		}
	}
	return true, label
}

func (c * comm) Read(ch chan<- []byte, label string) (int, bool){
	// c.conn.SetReadDeadline(time.Now().Add(time.Duration(READDEADLINE) * time.Second))
	// for debug
	stepCount++
	step += strconv.Itoa(stepCount) + " >> " + label + "\n"
	// for debug
	buff := make([]byte, TCPREADSIZE)
	n, err := c.conn.Read(buff)
	if err != nil {
		fmt.Println(step)
		fmt.Printf("Main Fail:\"Message Read Error\", Inner Fail:\"%v\", Error String: %v\n", label, err.Error())
		return n, false
	}else{
		ch <- buff[:n]
		return n, true
	}
	fmt.Println(step)
	fmt.Printf("Main Fail:\"Message Read Error\", Inner Fail:\"%v\", Error String: %v\n", label, err.Error())
	return n, false
}

func (c * comm) Close() bool{
 	err := c.conn.Close()
 	if err != nil {
		fmt.Printf("Main Fail:\"Comminication Close Error\", Error String: %v\n", err.Error())
 		return false
 	}
 	return true
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
	speed := int64(0)
	if int(fileSize) >= SPEEDTESTLIMIT {
		// run speed test
		res = com.SendTestData(intChan)
		if !res {return}else{speed = int64(<- intChan)}

	}else{
		// ack
		res = com.Ack()
		if !res {return}
		speed = int64(STANDARTSPEED)
	}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	// send filesize
	res = com.SendInt(speed)
	if !res {return}

	// receive ack
	res = com.Rec(boolChan)
	if !res {return}else{<-boolChan}

	batchSize := utils.GetPackNumber(fileSize, int64(READDISCBUFFER))
	innerBatchSize := utils.GetPackNumber(int64(READDISCBUFFER), speed)

	// // total := fileSize
	data := make([]byte, speed)
	datalen := int64(READDISCBUFFER)

	for i := 0 ; i < batchSize ; i++ {
		if *control == 1{
			if i == batchSize - 1 {
				data = rw.ReadAt(dir, off, fileSize - off)
				datalen = fileSize - off
				innerBatchSize = utils.GetPackNumber(datalen, speed)
			}else{
				data = rw.ReadAt(dir, off, int64(READDISCBUFFER))
			}
			innerData := make([]byte, 0, speed)
			for j := 0 ; j < innerBatchSize ; j++ {
				if datalen > speed {
					if j == innerBatchSize - 1{
						innerData = data[int(speed) * (j):]
						off += int64(len(innerData))
					}else{
						innerData = data[int(speed) * (j):int(speed) * (j+ 1)]
						off += int64(speed)
					}
				}else{
					innerData = data
					off += int64(len(innerData))
				}
				start := time.Now()
				res = com.SendData(innerData)
				end := time.Now()
				if !res {*control = 0;break}

				currentSpeed := float64(speed) / float64(end.Sub(start).Seconds() * 1e3) // kb/second 

				msg := fmt.Sprintf(`{"event":"prg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
				 username, ip, mac, port, id, dir, fileSize, off, int(currentSpeed))

				SendMsg(myIP, SRLISTENPORT, msg)
			}
		}else{

			msg := fmt.Sprintf(`{"event":"fprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
			 username, ip, mac, port, id, dir, fileSize, off, 0)
			SendMsg(myIP, MAINCOMANDPORT, msg)
			time.Sleep(10 * time.Millisecond)
			SendMsg(myIP, SRLISTENPORT, msg)
			com.Close()
			return
		}
	}
	msg := fmt.Sprintf(`{"event":"dprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"upload"}`,
	 username, ip, mac, port, id, dir, fileSize, off, 0)
	SendMsg(myIP, MAINCOMANDPORT, msg)
	time.Sleep(10 * time.Millisecond)
	SendMsg(myIP, SRLISTENPORT, msg)
	com.Close()
}


func ReceiveFile(ip, mac, username string, port int, id string, control * int){

	byteChan := make(chan []byte, 1)
	boolChan := make(chan bool, 1)
	int64Chan := make(chan int64, 1)

	// main comminication struct
	com := NewCom(ip, port)

	// listen start
	res := com.Listen()
    if !res{return}

    // receive file destination
    res = com.RecData(byteChan)
    if !res {return}

    dest := string(<- byteChan)

    // ack
    res = com.Ack()
    if !res{return}

    // receive filename
    res = com.RecData(byteChan)
    if !res {return}

    filename := string(<- byteChan)

    // ack
    res = com.Ack()
    if !res{return}

    // receive file size
    res = com.RecInt(int64Chan)
    if !res {return}
	
	fileSize := <-int64Chan

    filename = utils.UniqName(dest, filename, fileSize)
    dir := dest + utils.Sep + filename
	// offset := utils.GetFileSize(dir)
	// remaining := fileSize - offset

    // send offset of file
    // res = com.SendInt(0)
    // if !res {return}

    // if filesize bigger than speed test limit run a speed test
    if int(remaining) >= SPEEDTESTLIMIT {
        res = com.RecTestData()
        if !res {return}
    }else{
        res = com.Rec(boolChan)
        if !res{return}else{<-boolChan}
    }

    // ack
    res = com.Ack()
    if !res{return}

    // receive speed
    res = com.RecInt(int64Chan)
    if !res {return}
	
	speed := <-int64Chan

    // ack
    res = com.Ack()
    if !res{return}	

   // download file
    count := fileSize
    currentSize := int64(0)
    mainBuffer := make([]byte, 0, WRITEDISCBUFFER)
    genCount := 0
    currentSpeed := float64(0)
	printCount := int64(0)
    start := time.Now()
    for count > 0 {
    	if *control == 1 {
	        n, res := com.Read(byteChan, "Receive File Data")
	        if res {
	            mainBuffer = append(mainBuffer, (<-byteChan)...)
	        }else{*control = 0}
		
	        if genCount > WRITEDISCBUFFER{
	            rw.Write(dir, mainBuffer)
	            mainBuffer = make([]byte, 0, WRITEDISCBUFFER )
	            genCount = 0
	        }
			if printCount > speed{
		    	end := time.Now()
				currentSpeed = float64(printCount) / float64(end.Sub(start).Seconds() * 1e3)
				msg := fmt.Sprintf(`{"event":"prg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
				 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
				SendMsg(myIP, SRLISTENPORT, msg)
		    	start = end
				printCount = 0
			}			
	        count -= int64(n)
	        genCount += n
			printCount += int64(n)
    		currentSize = fileSize - count
    	}else{
			msg := fmt.Sprintf(`{"event":"fprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
			 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
			SendMsg(myIP, MAINCOMANDPORT, msg)
			time.Sleep(10 * time.Millisecond)
			SendMsg(myIP, SRLISTENPORT, msg)
			com.Close()
			done := rw.DelFile(dir)
			if done {
				SendMsg(myIP, SRLISTENPORT, fmt.Sprintf(`{"event":"info","content":"Unfinished file deleted. directory:%v"}`, dir))
			}else{
				SendMsg(myIP, SRLISTENPORT, fmt.Sprintf(`{"event":"info","content":"Unfinished file delete operation fail. directory:%v"}`, dir))
			}
			return
    	}
    }
    if len(mainBuffer) > 0 {
        rw.Write(dir, mainBuffer)
    }
	msg := fmt.Sprintf(`{"event":"dprg","username":"%v","ip":"%v","mac":"%v","port":"%v","uuid":"%v","dir":"%v","total":"%v","current":"%v","speed":"%v","type":"download"}`,
	 username, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed))
	SendMsg(myIP, MAINCOMANDPORT, msg)
	time.Sleep(10 * time.Millisecond)
	SendMsg(myIP, SRLISTENPORT, msg)
    com.Close()
}
// Receive bool , for ack and nack functions
func (c * comm) Rec(ch chan<- bool) bool{
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan, "Receive Ack/Nack")
	if !res {return false}
	byt := (<- byteChan)[0]
	if byt == 1 {
		ch <- true
	}else{
		ch <- false
	}
	return true
}

// send acknowledge
func (c * comm) Ack() bool{
	return c.Write([]byte{1}, "Send Ack")
}

// send not acknowledge
func (c * comm) Nack() bool{
	return c.Write([]byte{0}, "Send Nack")
}

// receive int64
func (c *comm) RecInt(int64Chan chan<- int64) bool{
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan, "Receive INT64")
	if res {
		numInt, _ := strconv.Atoi(string(<-byteChan))
		num := int64(numInt)
		int64Chan <- num
		return true
	}else{
		int64Chan <- int64(-1)
	}
	return false
}

// send int64
func (c *comm) SendInt(number int64) bool{
	strNum := strconv.FormatInt(number, 10)
	data := []byte(strNum)
    res := c.Write(data, "Send INT64")
    return res
}

// receive byte array
func (c *comm) RecData(chbyte chan<- []byte) bool{
	_, res := c.Read(chbyte, "Receive Data")
	return res
}

// send byte array
func (c * comm) SendData(data []byte) bool{
    return c.Write(data, "Send Data")
}

// Receive speed test data
func (c *comm) RecTestData() bool{
	byteChan := make(chan []byte, 1)
	count := MB
	for count > 0 {
		n, res := c.Read(byteChan, "Receive Test Data")
		if !res{return false}
		count -= n
		<-byteChan
	}
	return true
}

// Send speed test data
func (c *comm) SendTestData(ch chan<- int) bool{
	data := make([]byte, MB)
	res := false
	start := time.Now()
	res = c.SendData(data)
	if !res {return false}
	end := time.Now()
	ch <- int(float64(MB) / end.Sub(start).Seconds())
	return true
}

func (c * comm) Dial() bool{
	strPort := strconv.Itoa(c.port)
	conn, err := net.Dial("tcp", c.ip + ":" + strPort)
	if err != nil {
		fmt.Printf("Main Fail:\"Dial Connection\", Error String: %v\n", err.Error())
		return false
	}else{
		conn.SetReadDeadline(time.Now().Add(time.Duration(TCPDEADLINE) * time.Second))
	    c.conn = conn
	    return true
	}
}

func (c * comm) Listen() bool{
    strPort := strconv.Itoa(c.port)
    addr := myIP + ":" + strPort
    tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
    if err != nil {
		fmt.Printf("Main Fail:\"Adress Resolving\", Error String: %v\n", err.Error())
		return false
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
		fmt.Printf("Main Fail:\"Listen Error\", Error String: %v\n", err.Error())
		return false
    }else{
    	listener.SetDeadline(time.Now().Add(time.Duration(TCPDEADLINE) * time.Second))
	    conn, err := listener.Accept()
	    if err != nil {
			fmt.Printf("Main Fail:\"Listen Accept Error\", Error String: %v\n", err.Error())
			return false
	    }else{
	    	c.conn = conn
	    	return true
	    }
    }
}

func (c * comm) Write(msg []byte, label string) bool{
	// for debug
	stepCount++
	step += strconv.Itoa(stepCount) + " >> " + label + "\n"
	innerlabel := ""
	// for debug

	// FOR VM
	time.Sleep(100 * time.Millisecond)
	res := false
	for i := 0 ; i < WRITEREPETITION ; i ++ {
		// mini delay for try again
		time.Sleep(time.Duration(WRITEREPETITIONDELAY) * time.Millisecond)
		res, innerlabel = c.writeCore(msg)
		if res {
			return true
		}
	}
	// for debug
	fmt.Println(step)
	// for debug
	fmt.Printf("Main Fail:\"%v\", Inner Fail:\"%v\", Write Repetition:%v\n", innerlabel, label, WRITEREPETITION)
	return false
}

func (c * comm) writeCore(data []byte) (bool, string){
    label := ""
	n, err := c.conn.Write(data)
	if err != nil {
		// for debug
    	label = "Dial Connection " + err.Error()
    	return false, label
	}else{
		if n != len(data){
			return false, label
		}
	}
	return true, label
}

func (c * comm) Read(ch chan<- []byte, label string) (int, bool){
	// c.conn.SetReadDeadline(time.Now().Add(time.Duration(READDEADLINE) * time.Second))
	// for debug
	stepCount++
	step += strconv.Itoa(stepCount) + " >> " + label + "\n"
	// for debug
	buff := make([]byte, TCPREADSIZE)
	n, err := c.conn.Read(buff)
	if err != nil {
		fmt.Println(step)
		fmt.Printf("Main Fail:\"Message Read Error\", Inner Fail:\"%v\", Error String: %v\n", label, err.Error())
		return n, false
	}else{
		ch <- buff[:n]
		return n, true
	}
	fmt.Println(step)
	fmt.Printf("Main Fail:\"Message Read Error\", Inner Fail:\"%v\", Error String: %v\n", label, err.Error())
	return n, false
}

func (c * comm) Close() bool{
 	err := c.conn.Close()
 	if err != nil {
		fmt.Printf("Main Fail:\"Comminication Close Error\", Error String: %v\n", err.Error())
 		return false
 	}
 	return true
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