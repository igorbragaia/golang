package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

var err string
var myPort string 
var nServers int 
var CliConn []*net.UDPConn
var ServConn *net.UDPConn 

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
    for {
        n, addr, err := ServConn.ReadFromUDP(buf)
        fmt.Println("Received ", string(buf[0:n]), " from ", addr)
 
        if err != nil {
            fmt.Println("Error: ",err)
        } 
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
	myPort = os.Args[1]
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

func main() {
	initConnections()
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	
	i := 0
	for {
		go doServerJob()
		for j := 0; j < nServers; j++ {
			go doClientJob(j, i)
		}
		time.Sleep(time.Second * 1)
		i++
	}
}