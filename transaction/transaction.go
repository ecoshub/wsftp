package transaction

import (
	"github.com/ecoshub/jint"
	"github.com/ecoshub/penman"
	"net"
	"strconv"
	"time"
	"wsftp/tools"
)

const (
	ERROR_ADDRESS_RESOLVING     string = "Transaction: TCP IP resolve error."
	ERROR_CONNECTION_FAILED     string = "Transaction: TCP Connection error."
	ERROR_TCP_LISTEN_FAILED     string = "Transaction: TCP Listen error."
	ERROR_LISTEN_ACCECPT_FAILED string = "Transaction: TCP Listen accept error."
	ERROR_TCP_WRITE             string = "Transaction: TCP write error."
	ERROR_TCP_DIAL              string = "Transaction: TCP dial error."
	ERROR_TCP_READ              string = "Transaction: TCP read error."
	ERROR_CLOSE_CONN            string = "Transaction: TCP close connection error."

	WS_COMMANDER_LISTEN_PORT    string = "9999"
	WS_SEND_RECEIVE_LISTEN_PORT string = "10001"
	WS_MESSAGE_LISTEN_PORT      string = "10002"
)

var (
	INFO_FILE_DELETION      []byte = jint.MakeJson([]string{"event", "content"}, []string{"info", "Transaction: Unfinished file deleted."})
	INFO_FILE_DELETION_FAIL []byte = jint.MakeJson([]string{"event", "content"}, []string{"info", "Transaction: Unfinished file delete operation fail."})

	PROGRESS_SCHEME *jint.Scheme = jint.MakeScheme("event", "username", "nick", "ip", "mac", "port", "uuid", "dir", "total", "current", "speed", "type")

	MY_IP string = tools.MY_IP

	MB               int = 1048576
	// after debug set to 10 MB
	SPEED_TEST_LIMIT int = 1000 * MB
	STANDART_SPEED   int = 10 * MB
	TCPREADSIZE      int = 65536

	// write RAM to DISK tresholds
	WRITE_DISC_BUFFER int = 10 * MB
	READ_DISC_BUFFER  int = 10 * MB

	// general settings
	WRITE_REPETITION       int = 3
	WRITE_REPETITION_DELAY int = 10 // as ms
	TCP_DEADLINE           int = 5  // as second
)

// master comminication struct
type comm struct {
	ip   string
	port int
	conn net.Conn
}

func NewCom(ip string, port int) *comm {
	tempCon := comm{ip: ip, port: port}
	return &tempCon
}

// Receive bool , for ack and nack functions
func (c *comm) Rec(ch chan<- bool) bool {
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan)
	if !res {
		return false
	}
	byt := (<-byteChan)[0]
	if byt == 1 {
		ch <- true
	} else {
		ch <- false
	}
	return true
}

// send acknowledge
func (c *comm) Ack() bool {
	return c.Write([]byte{1})
}

// send not acknowledge
func (c *comm) Nack() bool {
	return c.Write([]byte{0})
}

// receive int64
func (c *comm) RecInt(int64Chan chan<- int64) bool {
	byteChan := make(chan []byte, 1)
	_, res := c.Read(byteChan)
	if res {
		numInt, _ := strconv.Atoi(string(<-byteChan))
		num := int64(numInt)
		int64Chan <- num
		return true
	} else {
		int64Chan <- int64(-1)
	}
	return false
}

// send int64
func (c *comm) SendInt(number int64) bool {
	strNum := strconv.FormatInt(number, 10)
	data := []byte(strNum)
	res := c.Write(data)
	return res
}

// receive byte array
func (c *comm) RecData(chbyte chan<- []byte) bool {
	_, res := c.Read(chbyte)
	return res
}

// send byte array
func (c *comm) SendData(data []byte) bool {
	return c.Write(data)
}

// Receive speed test data
func (c *comm) RecTestData() bool {
	byteChan := make(chan []byte, 1)
	count := MB
	for count > 0 {
		n, res := c.Read(byteChan)
		if !res {
			return false
		}
		count -= n
		<-byteChan
	}
	return true
}

// Send speed test data
func (c *comm) SendTestData(ch chan<- int) bool {
	data := make([]byte, MB)
	res := false
	start := time.Now()
	res = c.SendData(data)
	if !res {
		return false
	}
	end := time.Now()
	ch <- int(float64(MB) / end.Sub(start).Seconds())
	return true
}

func (c *comm) Dial() bool {
	strPort := strconv.Itoa(c.port)
	conn, err := net.Dial("tcp", c.ip+":"+strPort)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_CONNECTION_FAILED, err)
		return false
	} else {
		conn.SetReadDeadline(time.Now().Add(time.Duration(TCP_DEADLINE) * time.Second))
		c.conn = conn
		return true
	}
}

func (c *comm) Listen() bool {
	strPort := strconv.Itoa(c.port)
	addr := tools.MY_IP + ":" + strPort
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_ADDRESS_RESOLVING, err)
		return false
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_TCP_LISTEN_FAILED, err)
		return false
	} else {
		listener.SetDeadline(time.Now().Add(time.Duration(TCP_DEADLINE) * time.Second))
		conn, err := listener.Accept()
		if err != nil {
			tools.StdoutHandle("warning", ERROR_LISTEN_ACCECPT_FAILED, err)
			return false
		} else {
			c.conn = conn
			return true
		}
	}
}

func (c *comm) Write(msg []byte) bool {
	res := false
	for i := 0; i < WRITE_REPETITION; i++ {
		time.Sleep(time.Duration(WRITE_REPETITION_DELAY) * time.Millisecond)
		res = c.writeCore(msg)
		if res {
			return true
		}
	}
	tools.StdoutHandle("warning", ERROR_TCP_WRITE, nil)
	return false
}

func (c *comm) writeCore(data []byte) bool {
	n, err := c.conn.Write(data)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_TCP_DIAL, err)
		return false
	} else {
		if n != len(data) {
			return false
		}
	}
	return true
}

func (c *comm) Read(ch chan<- []byte) (int, bool) {
	buff := make([]byte, TCPREADSIZE)
	n, err := c.conn.Read(buff)
	if err != nil {
		tools.StdoutHandle("warning", ERROR_TCP_READ, err)
		return n, false
	}
	ch <- buff[:n]
	return n, true
}

func (c *comm) Close() bool {
	err := c.conn.Close()
	if err != nil {
		tools.StdoutHandle("warning", ERROR_CLOSE_CONN, err)
		return false
	}
	return true
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

func SendFile(ip, mac, username, nick string, port int, id, dir, dest string, control *int) {
	boolChan := make(chan bool, 1)
	intChan := make(chan int, 1)
	fmt.Println("send start", ip, port)
	// main comminication struct
	com := NewCom(ip, port)

	fileSize := tools.GetFileSize(dir)
	filename := tools.GetFileName(dir)

	fmt.Println("send", 1)
	// dial to receiver
	res := com.Dial()
	if !res {
		// *control = 0
		return
	}

	fmt.Println("send", 2)
	// send dest
	res = com.SendData([]byte(dest))
	if !res {
		return
	}

	fmt.Println("send", 3)
	// receive ack
	res = com.Rec(boolChan)
	if !res {
		return
	} else {
		<-boolChan
	}

	fmt.Println("send", 4)
	// send filename
	res = com.SendData([]byte(filename))
	if !res {
		return
	}

	fmt.Println("send", 5)
	// receive ack
	res = com.Rec(boolChan)
	if !res {
		return
	} else {
		<-boolChan
	}

	fmt.Println("send", 6)
	// send filesize
	res = com.SendInt(fileSize)
	if !res {
		return
	}

	fmt.Println("send", 7)
	// receive ack
	res = com.Rec(boolChan)
	if !res {
		return
	} else {
		<-boolChan
	}

	fmt.Println("send", 8)
	speed := int64(0)
	if int(fileSize) >= SPEED_TEST_LIMIT {
		// run speed test
		res = com.SendTestData(intChan)
		if !res {
			return
		} else {
			speed = int64(<-intChan)
		}
	} else {
		// ack
		res = com.Ack()
		if !res {
			return
		}
		speed = int64(STANDART_SPEED)
	}

	fmt.Println("send", 9)
	// receive ack
	res = com.Rec(boolChan)
	if !res {
		return
	} else {
		<-boolChan
	}

	fmt.Println("send", 10)
	// send filesize
	res = com.SendInt(speed)
	if !res {
		return
	}

	fmt.Println("send", 11)
	// receive ack
	res = com.Rec(boolChan)
	if !res {
		return
	} else {
		<-boolChan
	}

	fmt.Println("send", 12)
	batchSize := tools.GetPackNumber(fileSize, int64(READ_DISC_BUFFER))
	innerBatchSize := tools.GetPackNumber(int64(READ_DISC_BUFFER), speed)

	// // total := fileSize
	data := make([]byte, speed)
	datalen := int64(READ_DISC_BUFFER)

	off := int64(0)
	for i := 0; i < batchSize; i++ {
		if *control == 1 {
			if i == batchSize-1 {
				data = penman.ReadAt(dir, off, fileSize-off)
				datalen = fileSize - off
				innerBatchSize = tools.GetPackNumber(datalen, speed)
			} else {
				data = penman.ReadAt(dir, off, int64(READ_DISC_BUFFER))
			}
			innerData := make([]byte, 0, speed)
			for j := 0; j < innerBatchSize; j++ {
				if datalen > speed {
					if j == innerBatchSize-1 {
						innerData = data[int(speed)*(j):]
						off += int64(len(innerData))
					} else {
						innerData = data[int(speed)*(j) : int(speed)*(j+1)]
						off += int64(speed)
					}
				} else {
					innerData = data
					off += int64(len(innerData))
				}
				start := time.Now()
				res = com.SendData(innerData)
				end := time.Now()
				if !res {
					*control = 0
					break
				}
				currentSpeed := float64(speed) / float64(end.Sub(start).Seconds()*1e3) // kb/second
				progress := PROGRESS_SCHEME.MakeJson("prg", username, nick, ip, mac, port, id, dir, fileSize, off, int(currentSpeed), "upload")
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
			}
		} else {
			progress := PROGRESS_SCHEME.MakeJson("fprg", username, nick, ip, mac, port, id, dir, fileSize, off, 0, "upload")
			sendCore(MY_IP, WS_COMMANDER_LISTEN_PORT, progress)
			time.Sleep(10 * time.Millisecond)
			sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
			com.Close()
			return
		}
	}
	progress := PROGRESS_SCHEME.MakeJson("dprg", username, nick, ip, mac, port, id, dir, fileSize, off, 0, "upload")
	sendCore(MY_IP, WS_COMMANDER_LISTEN_PORT, progress)
	time.Sleep(10 * time.Millisecond)
	sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
	com.Close()
}

func ReceiveFile(ip, mac, username, nick string, port int, id string, control *int) {
	byteChan := make(chan []byte, 1)
	boolChan := make(chan bool, 1)
	int64Chan := make(chan int64, 1)

	fmt.Println("rec start", ip, port)

	// main comminication struct
	com := NewCom(ip, port)

	fmt.Println("rece", 1)
	// listen start
	res := com.Listen()
	if !res {
		return
	}

	fmt.Println("rece", 2)
	// receive file dest
	res = com.RecData(byteChan)
	if !res {
		return
	}

	fmt.Println("rece", 3)
	dest := string(<-byteChan)

	// ack
	res = com.Ack()
	if !res {
		return
	}

	fmt.Println("rece", 4)
	// receive filename
	res = com.RecData(byteChan)
	if !res {
		return
	}

	filename := string(<-byteChan)

	fmt.Println("rece", 5)
	// ack
	res = com.Ack()
	if !res {
		return
	}

	fmt.Println("rece", 6)
	// receive file size
	res = com.RecInt(int64Chan)
	if !res {
		return
	}

	fmt.Println("rece", 7)
	fileSize := <-int64Chan
	filename = tools.UniqName(dest, filename, fileSize)
	dir := dest + tools.SEPARATOR + filename

	fmt.Println("rece", 8)
	// ack
	res = com.Ack()
	if !res {
		return
	}

	fmt.Println("rece", 9)
	// if filesize bigger than speed test limit run a speed test
	if int(fileSize) >= SPEED_TEST_LIMIT {
		res = com.RecTestData()
		if !res {
			return
		}
	} else {
		res = com.Rec(boolChan)
		if !res {
			return
		} else {
			<-boolChan
		}
	}

	fmt.Println("rece", 10)
	// ack
	res = com.Ack()
	if !res {
		return
	}

	fmt.Println("rece", 11)
	// receive speed
	res = com.RecInt(int64Chan)
	if !res {
		return
	}

	fmt.Println("rece", 12)
	speed := <-int64Chan

	// ack
	res = com.Ack()
	if !res {
		return
	}

	fmt.Println("rece", 13)
	// download file
	count := fileSize
	currentSize := int64(0)
	mainBuffer := make([]byte, 0, WRITE_DISC_BUFFER)
	genCount := 0
	currentSpeed := float64(0)
	printCount := int64(0)
	start := time.Now()
	for count > 0 {
		if *control == 1 {
			n, res := com.Read(byteChan)
			if res {
				mainBuffer = append(mainBuffer, (<-byteChan)...)
			} else {
				*control = 0
			}

			if genCount > WRITE_DISC_BUFFER {
				penman.Write(dir, mainBuffer)
				mainBuffer = make([]byte, 0, WRITE_DISC_BUFFER)
				genCount = 0
			}
			if printCount > speed {
				end := time.Now()
				currentSpeed = float64(printCount) / float64(end.Sub(start).Seconds()*1e3)
				start = end
				progress := PROGRESS_SCHEME.MakeJson("prg", username, nick, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed), "download")
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
				printCount = 0
			}
			count -= int64(n)
			genCount += n
			printCount += int64(n)
			currentSize = fileSize - count
		} else {
			progress := PROGRESS_SCHEME.MakeJson("fprg", username, nick, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed), "download")
			sendCore(MY_IP, WS_COMMANDER_LISTEN_PORT, progress)
			time.Sleep(10 * time.Millisecond)
			sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
			com.Close()
			done := penman.DelFile(dir)
			if done {
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, INFO_FILE_DELETION)
			} else {
				sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, INFO_FILE_DELETION_FAIL)
			}
			return
		}
	}
	if len(mainBuffer) > 0 {
		penman.Write(dir, mainBuffer)
	}
	progress := PROGRESS_SCHEME.MakeJson("dprg", username, nick, ip, mac, port, id, dir, fileSize, currentSize, int(currentSpeed), "download")
	sendCore(MY_IP, WS_COMMANDER_LISTEN_PORT, progress)
	time.Sleep(10 * time.Millisecond)
	sendCore(MY_IP, WS_SEND_RECEIVE_LISTEN_PORT, progress)
	com.Close()
}
