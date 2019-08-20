package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"bufio"
)

var err string
var myPort string 
var nServers int 
var CliConn []*net.UDPConn
var ServConn *net.UDPConn 
var logicalClock int

func CheckError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func doServerJob() {
	buf := make([]byte, 1024)
	n, addr, err := ServConn.ReadFromUDP(buf)
	logicalClock_msg, err := strconv.Atoi(string(buf[0:n]))

	if logicalClock < logicalClock_msg {
		logicalClock = logicalClock_msg	
	}
	logicalClock++

	fmt.Printf("Received %s from %s\nCurrent Logical Clock = %d\n", string(buf[0:n]), addr, logicalClock)

	if err != nil {
		fmt.Println("Error: ",err)
	} 
}

func doClientJob(otherProcess int, i int) {
	Conn := CliConn[otherProcess]
	msg := strconv.Itoa(i)
	i++
	buf := []byte(msg)
	_,err := Conn.Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
}

func initConnections() {
	myPortId, err := strconv.Atoi(os.Args[1])
	myPort = os.Args[myPortId+1]
	nServers = len(os.Args) - 2

    ServerAddr, err := net.ResolveUDPAddr("udp",myPort);
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
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	ch := make(chan string)
	go readInput(ch)
	logicalClock = 0
	
	for {
		go doServerJob()
		select {
			case x, valid := <-ch:
				if valid {
					i1, err := strconv.Atoi(x)
					if (err == nil && i1 < len(os.Args) - 1 ){
						fmt.Printf("Notificar porta %s\n", os.Args[i1+1])
						go doClientJob(i1-1, logicalClock)
					} else {
						fmt.Println("Numero invalido")
					}
				} else {
					fmt.Println("Channel closed!")
				}
			default:
				time.Sleep(time.Second * 1)
		}
	}
}