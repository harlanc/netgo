package server

import (
	fmt "fmt"

	//"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/proto"
	"github.com/harlanc/netgo/lib"
)

//PeerStatus status of Peer
type PeerStatus uint32

const (
	_ PeerStatus = iota
	//Success status 1
	Success
	//CreateRoomSuccess status 2
	CreateRoomSuccess
	//CreateRoomAlreadyExist the room is exist already. 3
	CreateRoomAlreadyExist
	//JoinRoomSuccess status 4
	JoinRoomSuccess
	//JoinRoomNotExist the room is not exist. 5
	JoinRoomNotExist
	//JoinRoomAlreadyInRoom already in the room. 6
	JoinRoomAlreadyInRoom
	//JoinRoomFull the room is full already. 7
	JoinRoomFull
	//LeaveRoomSuccess leave a room success. 8
	LeaveRoomSuccess
)

//Peer struct
type Peer struct {
	//peer name
	Name string
	//peer ID ,it must be unique in a room
	ID uint32
	//the id of room in which current peer is
	RoomID string

	// //store the channel ids to which current peer send msg to .
	// //If it has no items,means that all the channels will receice
	// //it's messages.If the size is bigger than 0,only the channels in
	// //SendToMsgChannelsIds can receive the messages.All the channel ids
	// //in SendToMsgChannelsIds should be bigger than 0.
	// SendToMsgChannelsIds map[uint32]struct{}
	Conn *lib.Conn

	CachedInstantiationParams *InstantiationForwardParams
	CachedJoinRoomParams      *JoinRoomForwardParams
	//Map from unitID_methodname to *RPCForwardParams
	CachedRPCsMap map[string]*RPCForwardParams
}

//AsyncWriteMessage write message
func (peer *Peer) AsyncWriteMessage(msg proto.Message) error {

	responsemsg := msg
	if responsemsg == nil {
		fmt.Println("the response message should not be empty!")
	}

	bytes, err := proto.Marshal(responsemsg)
	//logger.LogInfo("first" + string(bytes[:]))
	if err != nil {
		fmt.Println(err)
	}

	//logger.LogInfo(lib.Int2String(len(bytes)))

	writepacket := NewNetgoPacket(uint32(len(bytes)), bytes)
	return peer.Conn.AsyncWritePacket(writepacket, 0)

}

//HandRequest hand a request
func (peer *Peer) HandRequest() {

}

//HandMessage hand a message from client
func (peer *Peer) HandMessage(packet lib.Packet) PeerStatus {
	// ipStr := peer.Conn.RemoteAddr().String()

	// defer func() {
	// 	fmt.Println(" Disconnected : " + ipStr)
	// 	peer.Conn.Close()
	// }()

	// buf := make([]byte, 4096, 4096)
	// for {
	// 	cnt, err := peer.Conn.Read(buf)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	p := packet.(*NetgoPacket)
	stReceive := &SendMessage{}

	body := p.GetBody()

	err := proto.Unmarshal(body, stReceive)
	if err != nil {
		panic(err)
	}

	//fmt.Println("receive", peer.Conn.GetRawConn().RemoteAddr(), stReceive)

	return peer.processMessage(stReceive, body)
	//}

}

func (peer *Peer) processMessage(msg *SendMessage, rawmsg []byte) PeerStatus {

	switch msg.MsgType {
	case MessageType_CreateRoom:
		return peer.CreateRoom(msg.CrParams.RoomId, msg.CrParams.MaxNumber)
	case MessageType_JoinRoom:
		return peer.JoinRoom(msg.JrParams.RoomId)
	case MessageType_JoinOrCreateRoom:
		return peer.JoinOrCreateRoom(msg.JocrParams.RoomId, msg.JocrParams.MaxNumber)
	case MessageType_LeaveRoom:
		return peer.LeaveRoom()
	case MessageType_RPC:
		return peer.RPC(msg.RpcParams)
	case MessageType_Instantiation:
		return peer.Instantiation(msg.IParams)
	case MessageType_ViewSync:
		return peer.ViewSync(msg.VsParams)
	case MessageType_CustomEvent:
		return peer.CustomEvent(msg.CeParams)
	case MessageType_SubscribeMsgChannels:
		return peer.SubscribeMsgChannels(msg.SmcParams.Channelids)
	case MessageType_UnSubscribeMsgChannels:
		return peer.UnSubscribeMsgChannels(msg.UsmcParams.Channelids)
	default:

	}
	return Success
}

//CreateRoom create a game room
func (peer *Peer) CreateRoom(roomid string, maxnumber uint32) PeerStatus {

	//must add a lock for ID2Room
	if _server.IsRoomExist(roomid) {
		return CreateRoomAlreadyExist
	}
	//create a new Room
	room := NewRoom(roomid, maxnumber)
	room.AddPeer(peer)

	_server.AddRoom(room)

	return CreateRoomSuccess
}

//JoinRoom join a game room
func (peer *Peer) JoinRoom(roomid string) PeerStatus {

	room, ok := _server.ID2Room[roomid]
	if !ok {
		return JoinRoomNotExist
	}

	if peer.ID != 0 {
		return JoinRoomAlreadyInRoom
	}

	if uint32(len(room.ID2Peer)) == room.MaxNumber {
		return JoinRoomFull
	}

	room.AddPeer(peer)

	var forwardParams JoinRoomForwardParams = JoinRoomForwardParams{}
	peer.CachedJoinRoomParams = &forwardParams

	var fmsg2others ForwardMessage
	fmsg2others.MsgType = MessageType_JoinRoom
	fmsg2others.JrfParams = &forwardParams
	fmsg2others.PeerId = peer.ID

	var recmsg2others ReceiveMessage
	recmsg2others.ReceiveMsgType = ReceiveMessageType_Forward
	recmsg2others.FMsg = &fmsg2others

	var fmsg2myself ForwardMessage
	fmsg2myself.MsgType = MessageType_JoinRoom
	// fmsg2myself.JrfParams = &forwardParams
	// fmsg2myself.PeerId = peer.ID

	var recmsg2myself ReceiveMessage
	recmsg2myself.ReceiveMsgType = ReceiveMessageType_Forward
	recmsg2myself.FMsg = &fmsg2myself

	peers := peer.GetAllSubscribePeers(nil)
	for _, v := range peers {
		if v.ID != peer.ID {
			v.AsyncWriteMessage(&recmsg2others)

			fmsg2myself.JrfParams = v.CachedJoinRoomParams
			fmsg2myself.PeerId = v.ID
			peer.AsyncWriteMessage(&recmsg2myself)
		}
	}

	return JoinRoomSuccess
}

//JoinOrCreateRoom if exist join else create room
func (peer *Peer) JoinOrCreateRoom(roomid string, maxnumber uint32) PeerStatus {

	_, err := _server.GetRoom(roomid)

	if err != nil {
		return peer.CreateRoom(roomid, maxnumber)
	}
	return peer.JoinRoom(roomid)
}

//LeaveRoom leave a game room
func (peer *Peer) LeaveRoom() PeerStatus {

	peers := peer.GetAllSubscribePeers(nil)

	room, ok := _server.ID2Room[peer.RoomID]
	if ok {
		//delete revelant peer from the Subscribed Channels.
		room.UnSubscribeAllChannels(peer.ID)
		//remove the Peer from current room
		room.DeletePeer(peer)

		var forwardParams LeaveRoomForwardParams = LeaveRoomForwardParams{}

		var forwardmsg ForwardMessage
		forwardmsg.MsgType = MessageType_LeaveRoom
		forwardmsg.LrfParams = &forwardParams
		forwardmsg.PeerId = peer.ID

		var recmessage ReceiveMessage
		recmessage.ReceiveMsgType = ReceiveMessageType_Forward
		recmessage.FMsg = &forwardmsg

		for _, v := range peers {
			if v.ID != peer.ID {
				v.AsyncWriteMessage(&recmessage)
			}
		}
	}
	if room.GetPeersCount() == 0 {
		_server.DeleteRoom(peer.RoomID)
	}
	return LeaveRoomSuccess
}

//RPC rpc method
func (peer *Peer) RPC(params *RPCParams) PeerStatus {

	peers := peer.GetAllSubscribePeers(params.Options)

	var forwardParams RPCForwardParams
	forwardParams.MethodName = params.MethodName
	forwardParams.ViewID = params.ViewID
	forwardParams.Parameters = params.Parameters

	var forwardmsg ForwardMessage
	forwardmsg.MsgType = MessageType_RPC
	forwardmsg.RfParams = &forwardParams
	forwardmsg.PeerId = peer.ID

	var recmessage ReceiveMessage
	recmessage.ReceiveMsgType = ReceiveMessageType_Forward
	recmessage.FMsg = &forwardmsg

	switch params.Target {

	case RPCTarget_All:
		for _, v := range peers {
			v.AsyncWriteMessage(&recmessage)
		}
	case RPCTarget_Others:

		for _, v := range peers {
			if v.ID != peer.ID {
				v.AsyncWriteMessage(&recmessage)
			}
		}
	}

	return Success
}

//Instantiation Instantiation a prefab
func (peer *Peer) Instantiation(params *InstantiationParams) PeerStatus {

	peers := peer.GetAllSubscribePeers(params.Options)

	var forwardParams InstantiationForwardParams

	forwardParams.PrefabName = params.PrefabName
	forwardParams.Position = params.Position
	forwardParams.Rotation = params.Rotation
	forwardParams.ViewIDs = params.ViewIDs

	//save local
	peer.CachedInstantiationParams = &forwardParams

	//let others' instantiate my prefab
	var forwardmsgothers ForwardMessage
	forwardmsgothers.MsgType = MessageType_Instantiation
	forwardmsgothers.IfParams = &forwardParams
	forwardmsgothers.PeerId = peer.ID

	var recmessageothers ReceiveMessage
	recmessageothers.ReceiveMsgType = ReceiveMessageType_Forward
	recmessageothers.FMsg = &forwardmsgothers

	//others' prefab instantiated in my room
	var forwardmsgmyself ForwardMessage
	forwardmsgmyself.MsgType = MessageType_Instantiation

	var recmessagemyself ReceiveMessage
	recmessagemyself.ReceiveMsgType = ReceiveMessageType_Forward
	recmessagemyself.FMsg = &forwardmsgmyself

	for _, v := range peers {
		if v.ID != peer.ID {

			v.AsyncWriteMessage(&recmessageothers)

			forwardmsgmyself.IfParams = v.CachedInstantiationParams
			forwardmsgmyself.PeerId = v.ID
			peer.AsyncWriteMessage(&recmessagemyself)
		}
	}
	return Success
}

//ViewSync Instantiation a prefab
func (peer *Peer) ViewSync(params *ViewSyncParams) PeerStatus {

	peers := peer.GetAllSubscribePeers(params.Options)

	var forwardParams ViewSyncForwardParams
	forwardParams.VsdParams = params.VsdParams

	var forwardmsg ForwardMessage
	forwardmsg.MsgType = MessageType_ViewSync
	forwardmsg.VsfParams = &forwardParams

	var recmessage ReceiveMessage
	recmessage.ReceiveMsgType = ReceiveMessageType_Forward
	recmessage.FMsg = &forwardmsg

	for _, v := range peers {
		if v.ID != peer.ID {
			v.AsyncWriteMessage(&recmessage)
		}
	}
	return Success
}

//CustomEvent issue custom event
func (peer *Peer) CustomEvent(params *CustomEventParams) PeerStatus {

	peers := peer.GetAllSubscribePeers(params.Options)

	var forwardParams CustomEventForwardParams
	forwardParams.EventID = params.EventID
	forwardParams.CustomData = params.CustomData

	var forwardmsg ForwardMessage
	forwardmsg.MsgType = MessageType_CustomEvent
	forwardmsg.CeParams = &forwardParams

	var recmessage ReceiveMessage
	recmessage.ReceiveMsgType = ReceiveMessageType_Forward
	recmessage.FMsg = &forwardmsg

	//TODO
	//targetids := params.TargetPeerIds

	for _, v := range peers {
		if v.ID != peer.ID {
			v.AsyncWriteMessage(&recmessage)
		}
	}
	return Success

}

//GetAllSubscribePeers get all the subscribe peers
func (peer *Peer) GetAllSubscribePeers(commonoptions *CommonOptions) []*Peer {

	var channels []uint32

	if commonoptions == nil {
		channels = make([]uint32, 0)
	} else {
		channels = commonoptions.SendToChannelIds
	}

	room, _ := _server.GetRoom(peer.RoomID)
	return room.GetAllSubscribePeers(channels)
}

//SubscribeMsgChannels subscribe msg channels
func (peer *Peer) SubscribeMsgChannels(channelids []uint32) PeerStatus {

	room, _ := _server.GetRoom(peer.RoomID)
	for _, v := range channelids {
		room.SubscribeChannel(v, peer.ID)
	}
	return Success
}

//UnSubscribeMsgChannels unsubscribe msg channels
func (peer *Peer) UnSubscribeMsgChannels(channelids []uint32) PeerStatus {

	room, _ := _server.GetRoom(peer.RoomID)
	for _, v := range channelids {
		room.UnSubscribeChannel(v, peer.ID)
	}
	return Success
}
