package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"bufio"
	"encoding/json"
)

type Message struct {
	Time int
	Processor int
	Text string
}

var state string
var counter int
var qty int
var err string
var myPortId int
var myPort string
var nServers int
var CliConn []*net.UDPConn
var ServConn *net.UDPConn
var ResourceConn *net.UDPConn
var logicalClock int
var logicalClockFreeze int
var queue []int

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func doServerJob() {
	buf := make([]byte, 1024)
	n, _, err := ServConn.ReadFromUDP(buf)
	var receivedMessage Message
	err = json.Unmarshal(buf[:n], &receivedMessage)
	if err != nil {
		fmt.Println("Unmarshal server response failed.")
	}

	if logicalClock < receivedMessage.Time {
		logicalClock = receivedMessage.Time
	}
	logicalClock++

	if receivedMessage.Text == "REPLY" {
		fmt.Printf("[logical clock %d] REPLY from %d\n", logicalClock, receivedMessage.Processor)
		counter++
		if counter == qty {
			state = "HELD"
			fmt.Printf("[logical clock %d] ENTROU NA CS\n", logicalClock)
			useResource()

			state = "RELEASED"
			counter = 0
			fmt.Printf("[logical clock %d] Replying to: ", logicalClock)
			fmt.Println(queue)
			for _, p := range queue {
				doClientJob(p-1, logicalClock, "REPLY")
			}
			queue = []int{}
			fmt.Printf("[logical clock %d] SAIU DA CS\n", logicalClock)
			fmt.Println("*************************************\n")
		}
	} else if receivedMessage.Text == "REQUEST" {
		if state == "HELD" || (state == "WANTED" && logicalClockFreeze < receivedMessage.Time) {
			queue = append(queue, receivedMessage.Processor)
		} else {
			fmt.Printf("[logical clock %d] REPLYING TO %d\n", logicalClock, receivedMessage.Processor)
			doClientJob(receivedMessage.Processor-1, logicalClock, "REPLY")
		}
	}

	if err != nil {
		fmt.Println("Error: ",err)
	}
}

func useResource() {
	msg := Message{
		Time: logicalClock,
		Processor: myPortId,
		Text: "I'm into CS",
	}

	jsonRequest, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Marshal connection information failed.")
	}

	_, err = ResourceConn.Write(jsonRequest)
	if err != nil {
		fmt.Println(jsonRequest, err)
	}

	time.Sleep(5 * time.Second)
}

func doClientJob(otherProcess int, logicalClock int, text string) {
	Conn := CliConn[otherProcess]

	msg := Message{
		Time: logicalClock,
		Processor: myPortId,
		Text: text,
	}

	jsonRequest, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Marshal connection information failed.")
	}

	_, err = Conn.Write(jsonRequest)
	if err != nil {
		fmt.Println(jsonRequest, err)
	}
}

func initConnections() {
	id, err := strconv.Atoi(os.Args[1])
	myPortId = id
	myPort = os.Args[myPortId+1]
	nServers = len(os.Args) - 2
	qty = len(os.Args) - 3

	ServerAddr, err := net.ResolveUDPAddr("udp",myPort)
	CheckError(err)
	Conn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)

	ServConn = Conn

	for i := 0; i < nServers; i++ {
		ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + os.Args[i+2])
		CheckError(err)
		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
		CheckError(err)
		Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
		CheckError(err)

		CliConn = append(CliConn, Conn)
	}

	ServerAddr,err = net.ResolveUDPAddr("udp","127.0.0.1:10001")
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)
	Conn, err = net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	ResourceConn = Conn
}

func readInput(ch chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func main() {
	initConnections()
	defer ServConn.Close()
	logicalClock = 1
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	state = "RELEASED"

	ch := make(chan string)
	go readInput(ch)

	for {
		go doServerJob()
		select {
		case text, valid := <-ch:
			if valid {
				if text == "x" {
					state = "WANTED"
					logicalClockFreeze = logicalClock
					for i := 0; i < nServers; i++ {
						if i != myPortId - 1 {
							go doClientJob(i, logicalClock, "REQUEST")
						}
					}
				}
			} else {
				fmt.Println("Channel closed!")
			}
		default:
			time.Sleep(time.Second * 1)
		}
	}
}