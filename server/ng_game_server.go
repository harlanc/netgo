package server

import (
	"container/heap"
	"errors"
	"strings"
	"sync"

	"github.com/harlanc/netgo/lib"
)

//MsgChannel message channels
type MsgChannel struct {
	ID uint32
	//peer ids that the current channel is subscribed
	PeerIds map[uint32]struct{}
	//test  sync.Map[int]string
}

//NewMsgChannel new a MsgChannel
func NewMsgChannel(id uint32) *MsgChannel {
	channel := MsgChannel{ID: id, PeerIds: make(map[uint32]struct{})}
	return &channel
}

//GameServer struct
type GameServer struct {
	ID        string
	PeerCount int
	//A map from Room Id to Room
	ID2Room map[string]*Room
	//add mux for server
	RoomIDMux sync.RWMutex
}

//GetRoom get room from server
func (server *GameServer) GetRoom(roomid string) (*Room, error) {

	server.RoomIDMux.RLock()
	defer func() {
		server.RoomIDMux.RUnlock()
	}()

	if server.ID2Room != nil {
		if room, ok := server.ID2Room[roomid]; ok {
			return room, nil
		}

	}

	return nil, errors.New("No room with ID" + roomid + " exists!")

}

func (server *GameServer) IsRoomExist(roomid string) bool {

	server.RoomIDMux.RLock()
	defer func() {
		server.RoomIDMux.RUnlock()
	}()

	if _, ok := server.ID2Room[roomid]; ok {
		return true
	}

	return false

}

func (server *GameServer) AddRoom(room *Room) {

	server.RoomIDMux.Lock()
	defer func() {
		server.RoomIDMux.Unlock()
	}()

	server.ID2Room[room.ID] = room

}

func (server *GameServer) DeleteRoom(roomid string) {

	server.RoomIDMux.Lock()
	defer func() {
		server.RoomIDMux.Unlock()
	}()

	delete(server.ID2Room, roomid)
}

func (server *GameServer) GetPeer(roomid string, peerid uint32) (error, *Peer) {
	room, err := server.GetRoom(roomid)
	if err != nil {
		return err, nil
	}

	err, peer := room.GetPeer(peerid)
	if err != nil {
		return err, nil
	}
	return nil, peer

}

var _server GameServer = GameServer{ID: "test", RoomIDMux: sync.RWMutex{}, ID2Room: map[string]*Room{}}

//Room Struct
type Room struct {
	ID        string
	MaxNumber uint32
	// A room can contain serval channels,and a client can listen to these channels,then it will
	// receive messages from these listened channels,channel 0 is a default channel which all the clients will
	// listen to.
	ID2MsgChannel map[uint32]*MsgChannel
	//Protect MsgChannel
	MsgChannelMux sync.RWMutex
	//A map from Peer Id to Peer
	ID2Peer map[uint32]*Peer
	//PeerMux protect the PeerIDPool
	PeerMux sync.RWMutex
	//PeerIDPool here we use a heap to store all the peer ids for a room,
	//we can get the smallest peer id when a new peer joins the room by
	//calling the func GetPeerID.
	PeerIDPool *lib.IntHeap
}

//GetAllSubscribePeers get all the subscribe peers
func (room *Room) GetAllSubscribePeers(sendToChannelsIds []uint32) []*Peer {

	var peers []*Peer

	if len(sendToChannelsIds) == 0 {
		return room.GetAllPeers()
	}

	room.MsgChannelMux.RLock()
	defer func() {
		room.MsgChannelMux.RUnlock()
	}()

	var peerids map[uint32]struct{} = make(map[uint32]struct{})

	for _, i := range sendToChannelsIds {
		for v := range room.ID2MsgChannel[i].PeerIds {
			if _, ok := peerids[v]; !ok {
				peerids[v] = struct{}{}
				_, peer := room.GetPeer(v)
				peers = append(peers, peer)
			}
		}
	}

	return peers
}

//SubscribeChannel a peer subscribe a channel
func (room *Room) SubscribeChannel(channelid uint32, peerid uint32) {

	room.MsgChannelMux.Lock()
	defer func() {
		room.MsgChannelMux.Unlock()
	}()

	if channel, ok := room.ID2MsgChannel[channelid]; ok {
		channel.PeerIds[peerid] = struct{}{}
	} else {
		//create a new channel and subscribe
		newchannel := NewMsgChannel(channelid)
		newchannel.PeerIds[peerid] = struct{}{}
		room.ID2MsgChannel[channelid] = newchannel
	}
}

//UnSubscribeChannel a peer unsubscribe a channel
func (room *Room) UnSubscribeChannel(channelid uint32, peerid uint32) {

	room.MsgChannelMux.Lock()
	defer func() {
		room.MsgChannelMux.Unlock()
	}()

	if channel, ok := room.ID2MsgChannel[channelid]; ok {
		delete(channel.PeerIds, peerid)
	}
}

//UnSubscribeAllChannels when a peer left room,this operation will be called
func (room *Room) UnSubscribeAllChannels(peerid uint32) {

	room.MsgChannelMux.Lock()
	defer func() {
		room.MsgChannelMux.Unlock()
	}()

	for _, v := range room.ID2MsgChannel {
		if _, ok := v.PeerIds[peerid]; ok {
			delete(v.PeerIds, peerid)
		}
	}
}

//GetPeersCount get the peer count in the room
func (room *Room) GetPeersCount() int {

	room.MsgChannelMux.RLock()
	defer func() {
		room.MsgChannelMux.RUnlock()
	}()

	return len(room.ID2Peer)
}

//GetAllPeers get all the peers from the specified room
func (room *Room) GetAllPeers() []*Peer {

	room.PeerMux.RLock()
	defer func() {
		room.PeerMux.RUnlock()
	}()

	var peers []*Peer

	for _, v := range room.ID2Peer {
		peers = append(peers, v)
	}
	return peers

}

//GetPeer get a specified peer
func (room *Room) GetPeer(peerid uint32) (error, *Peer) {

	room.PeerMux.RLock()
	defer func() {
		room.PeerMux.RUnlock()
	}()

	if peer, ok := room.ID2Peer[peerid]; ok {
		return nil, peer
	}

	return errors.New("No peer with ID" + lib.Uint322String(peerid) + " exists!"), nil

}

//AddPeer return the smallest peer id from the heap and then re-establish the heap.
func (room *Room) AddPeer(peer *Peer) {

	room.PeerMux.Lock()
	defer func() {
		room.PeerMux.Unlock()
	}()
	peerid := room.PeerIDPool.Pop()
	peer.ID = peerid.(uint32)

	room.ID2Peer[peer.ID] = peer
	peer.RoomID = room.ID
	peer.Conn.RoomID = room.ID
	peer.Conn.PeerID = peer.ID
}

//DeletePeer release the peer ID when someone leaves the room.
//push the peerid back to the heap.
func (room *Room) DeletePeer(peer *Peer) {

	room.PeerMux.Lock()
	defer func() {
		room.PeerMux.Unlock()
	}()
	peer.Conn.PeerID = 0
	heap.Push(room.PeerIDPool, peer.ID)
	delete(room.ID2Peer, peer.ID)
}

//NewRoom instantiation a room
func NewRoom(id string, maxnumber uint32) *Room {

	idPool := &lib.IntHeap{}
	for i := uint32(1); i <= maxnumber; i++ {
		idPool.Push(i)
	}
	heap.Init(idPool)

	room := Room{ID: id, MaxNumber: maxnumber, PeerMux: sync.RWMutex{}, PeerIDPool: idPool, ID2Peer: map[uint32]*Peer{}, MsgChannelMux: sync.RWMutex{}}
	return &room
}

//EmptyString a empty string
const EmptyString string = ""

//GetPeerWorldID Peer world id equals gameid + Peerid,used for communication between
//rooms from different game servers
func (game *GameServer) GetPeerWorldID(peer Peer) string {
	return strings.Join([]string{game.ID}, lib.Uint322String(peer.ID))
}
