package lib

import (
	"net"
)

//Packet interface
type Packet interface {
	Serialize() []byte
}

//Protocol interface
type Protocol interface {
	ReadPacket(conn *net.TCPConn) (Packet, error)
}
