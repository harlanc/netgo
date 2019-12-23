package server

import (
	"fmt"
	"net"
)

const ReadBufferSize int = 1024

func Listen(network, address string) {

	var tcpAddr *net.TCPAddr

	tcpAddr, _ = net.ResolveTCPAddr(network, address)

	tcpListener, _ := net.ListenTCP(network, tcpAddr)
	defer tcpListener.Close()
	fmt.Println("Server ready to read ...")

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		peer := Peer{Conn: nil}
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("A client connected :" + tcpConn.RemoteAddr().String())
		go peer.HandRequest()
	}
}

func ListenU(network, address string) {

	var udpAddr *net.UDPAddr
	udpAddr, _ = net.ResolveUDPAddr(network, address)

	udpSocket, err := net.ListenUDP(network, udpAddr)

	if err != nil {
		fmt.Println("Listen UDP failed")
		return
	}

	defer udpSocket.Close()
}
