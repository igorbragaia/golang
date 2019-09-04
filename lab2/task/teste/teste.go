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

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(0)
	}
}

func main() {
	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
	CheckError(err)
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	CheckError(err)
	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	CheckError(err)

	defer Conn.Close()

	msg := Message{
		Time:      1,
		Processor: 2,
		Text:      "abc",
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