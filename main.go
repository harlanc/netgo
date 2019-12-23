package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/harlanc/netgo/lib"
	"github.com/harlanc/netgo/server"
)

type Callback struct{}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// creates a tcp listener
	tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.0.104:8686")
	//tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.50.236:8686")
	//tcpAddr, err := net.ResolveTCPAddr("tcp", "192.168.43.189:8686")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	// creates a server
	config := &lib.Config{
		PacketSendChanLimit:    2048,
		PacketReceiveChanLimit: 2048,
	}
	srv := lib.NewServer(config, &server.NetgoCallback{}, &server.NetgoProtocol{})

	// starts service
	go srv.Start(listener, time.Second)
	fmt.Println("listening:", listener.Addr())

	// catchs system signal
	chSig := make(chan os.Signal)
	signal.Notify(chSig, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Signal: ", <-chSig)

	// stops service
	srv.Stop()
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
