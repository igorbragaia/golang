package main

import (
	"fmt"
	"net"
	"encoding/json"
	"os"
)

type Message struct {
	Time int
	Processor int
	Text string
}

var ServConn *net.UDPConn 
var port string

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
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

	fmt.Printf("[logical clock %d] Processor %d says %s\n", receivedMessage.Time, receivedMessage.Processor, receivedMessage.Text)

	if err != nil {
		fmt.Println("Error: ",err)
	} 
}

func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	ServConn = Connection
	CheckError(err)
	defer ServConn.Close()
	for {
		doServerJob()
	}
}