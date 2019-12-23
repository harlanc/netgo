package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/harlanc/netgo/lib"
	"github.com/harlanc/netgo/logger"
)

type NetgoPacket struct {
	length uint32
	body   []byte
}

func (this *NetgoPacket) Serialize() []byte {

	// logger.LogInfo(lib.Uint322String(this.length))
	// logger.LogInfo(string(this.body[:]))
	rv := make([]byte, 4+len(this.body))
	copy(rv, lib.Uint322ByteArray(this.length))

	//https://www.golangprograms.com/different-ways-to-convert-byte-array-into-string.html
	//logger.LogInfo(string(rv[:]))
	copy(rv[4:], this.body)
	//logger.LogInfo(string(rv[:]))
	return rv
}

// func (this *NetgoPacket) GetLength() uint32 {
// 	return binary.BigEndian.Uint32(this.buff[0:4])
// }

func (this *NetgoPacket) GetBody() []byte {
	return this.body
}

func NewNetgoPacket(len uint32, buff []byte) *NetgoPacket {
	p := &NetgoPacket{}

	p.length = len
	p.body = make([]byte, len)
	copy(p.body, buff)
	return p
}

type NetgoProtocol struct {
}

//ReadPacket read the packet from network
func (protocol *NetgoProtocol) ReadPacket(conn *net.TCPConn) (lib.Packet, error) {
	var (
		lengthBytes []byte = make([]byte, 4)
		length      uint32
	)

	// read length

	conn.SetReadDeadline(time.Now().Add(120 * time.Second))

	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}

	if length = binary.LittleEndian.Uint32(lengthBytes); length > 2048 {
		return nil, errors.New("the size of packet is larger than the limit" + lib.Uint322String(length))
	}

	body := make([]byte, length)
	// copy(buff[0:4], lengthBytes)

	// read body ( buff = lengthBytes + body )
	if _, err := io.ReadFull(conn, body); err != nil {
		return nil, err
	}

	return NewNetgoPacket(length, body), nil
}

//NetgoCallback struct
type NetgoCallback struct {
}

//OnConnect connect
func (callback *NetgoCallback) OnConnect(c *lib.Conn) bool {
	addr := c.GetRawConn().RemoteAddr()
	c.PutExtraData(addr)
	fmt.Println("OnConnect:", addr)
	peer := &Peer{Conn: c}
	var resssmsg ResponseSocketStatusMessage = ResponseSocketStatusMessage{}
	resssmsg.SStatus = SocketStatus_Connected

	var receivemsg ReceiveMessage = ReceiveMessage{}
	receivemsg.ReceiveMsgType = ReceiveMessageType_ResponseSocketStatus
	receivemsg.RssMsg = &resssmsg
	peer.AsyncWriteMessage(&receivemsg)
	//c.AsyncWritePacket(NewTelnetPacket("unknow", []byte("Welcome to this Telnet Server")), 0)
	return true
}

//OnMessage on message
func (callback *NetgoCallback) OnMessage(c *lib.Conn, p lib.Packet) bool {
	packet := p.(*NetgoPacket)

	var peer *Peer
	if c.PeerID == 0 {
		peer = &Peer{Conn: c}
	} else {

		var err error

		err, peer = _server.GetPeer(c.RoomID, c.PeerID)
		if err != nil {
			logger.LogError(err.Error())
			return false
		}

	}

	status := peer.HandMessage(packet)

	var resmsg ResponseOperationMessage

	switch status {
	case CreateRoomSuccess:
		params := CreateRoomResponseParams{ReturnValue: uint32(CreateRoomSuccess), PeerId: peer.ID}
		resmsg = ResponseOperationMessage{MsgType: MessageType_CreateRoom, CrrParams: &params}
	case CreateRoomAlreadyExist:
		params := CreateRoomResponseParams{ReturnValue: uint32(CreateRoomAlreadyExist)}
		resmsg = ResponseOperationMessage{MsgType: MessageType_CreateRoom, CrrParams: &params}
	case JoinRoomAlreadyInRoom:
		params := JoinRoomResponseParams{ReturnValue: uint32(JoinRoomAlreadyInRoom)}
		resmsg = ResponseOperationMessage{MsgType: MessageType_JoinRoom, JrrParams: &params}
	case JoinRoomFull:
		params := JoinRoomResponseParams{ReturnValue: uint32(JoinRoomFull)}
		resmsg = ResponseOperationMessage{MsgType: MessageType_JoinRoom, JrrParams: &params}
	case JoinRoomNotExist:
		params := JoinRoomResponseParams{ReturnValue: uint32(JoinRoomNotExist)}
		resmsg = ResponseOperationMessage{MsgType: MessageType_JoinRoom, JrrParams: &params}
	case JoinRoomSuccess:
		params := JoinRoomResponseParams{ReturnValue: uint32(JoinRoomSuccess), PeerId: peer.ID}
		resmsg = ResponseOperationMessage{MsgType: MessageType_JoinRoom, JrrParams: &params}
	case LeaveRoomSuccess:
		params := LeaveRoomResponseParams{ReturnValue: uint32(LeaveRoomSuccess), PeerId: peer.ID}
		resmsg = ResponseOperationMessage{MsgType: MessageType_LeaveRoom, LrrParams: &params}

	default:
		return true

	}

	var receivemsg ReceiveMessage = ReceiveMessage{}
	receivemsg.ReceiveMsgType = ReceiveMessageType_ResponseOperation
	receivemsg.RoMsg = &resmsg

	var rv = peer.AsyncWriteMessage(&receivemsg)
	if rv != nil {
		logger.LogError("AsyncWriteMessage error" + rv.Error())
		return false
	}

	return true
}

//OnClose on close
func (callback *NetgoCallback) OnClose(c *lib.Conn) {

	room, err := _server.GetRoom(c.RoomID)
	if err == nil {
		err, peer := room.GetPeer(c.PeerID)
		if err != nil {
			logger.LogError(err.Error())
			//return
		}
		peer.LeaveRoom()
	} else {
		logger.LogError(err.Error())
		//return
	}

	fmt.Println("OnClose:", c.GetExtraData())
}
