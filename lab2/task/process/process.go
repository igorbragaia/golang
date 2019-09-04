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

type ClockStruct struct {
	Id int
	Clocks []int
}

var err string
var myPortId int
var myPort string 
var nServers int 
var CliConn []*net.UDPConn
var ServConn *net.UDPConn 
var logicalClock ClockStruct

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
	var logicalClockMessage ClockStruct
	err = json.Unmarshal(buf[:n], &logicalClockMessage)
	if err != nil {
		fmt.Println("Unmarshal server response failed.")
	}

	logicalClock.Clocks[myPortId]++
	for i := 1; i < len(logicalClock.Clocks); i++ {
		if logicalClockMessage.Clocks[i] > logicalClock.Clocks[i] {
			logicalClock.Clocks[i] = logicalClockMessage.Clocks[i]
		}
	}

	fmt.Printf("Current Logical Clock = %d\n", logicalClock.Clocks[1:nServers+1])

	if err != nil {
		fmt.Println("Error: ",err)
	} 
}

func doClientJob(otherProcess int, i ClockStruct) {
	Conn := CliConn[otherProcess]

	jsonRequest, err := json.Marshal(logicalClock)
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
	logicalClock.Id = myPortId
	logicalClock.Clocks = append(logicalClock.Clocks, 0)
	for i := 0; i < nServers; i++ {
		logicalClock.Clocks = append(logicalClock.Clocks, 0)
		defer CliConn[i].Close()
	}

	ch := make(chan string)
	go readInput(ch)
	
	for {
		go doServerJob()
		select {
			case x, valid := <-ch:
				if valid {
					i1, err := strconv.Atoi(x)
					if (err == nil && i1 < len(os.Args) - 1 ){
						fmt.Printf("Notify port %s\n", os.Args[i1+1])
						go doClientJob(i1-1, logicalClock)
					} else {
						fmt.Println("Invalid number")
					}
				} else {
					fmt.Println("Channel closed!")
				}
			default:
				time.Sleep(time.Second * 1)
		}
	}
}