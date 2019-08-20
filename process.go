package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
//dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
//mensagens dos outros processos)

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
	//Ler (uma vez somente) da conexão UDP a mensagem
	//Escrever na tela a msg recebida (indicando o endereço de quem enviou)
	buf := make([]byte, 1024)
    for {
        n,addr,err := ServConn.ReadFromUDP(buf)
        fmt.Println("Received ",string(buf[0:n]), " from ",addr)
 
        if err != nil {
            fmt.Println("Error: ",err)
        } 
    }
}

func doClientJob(otherProcess int, i int) {
	//Enviar uma mensagem (com valor i) para o servidor do processo
	//otherServer
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
	
	//Todo Process fará a mesma coisa: ouvir msg e mandar infinitos i’s para os outros processos
	
	i := 0
	for {
		//Server
		go doServerJob()
		//Client
		for j := 0; j < nServers; j++ {
			go doClientJob(j, i)
		}
		// Wait a while
		time.Sleep(time.Second * 1)
		i++
	}
}