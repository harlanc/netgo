package server

type Lobby struct {
	Lobbyname string
	Gamelist  []GameServer
}

type LobbyOperations interface {
	JoinLobby()
	GetGameServers() []GameServer
}

//join a lobby
func (lobby *Lobby) JoinLobby() {

}

func (lobby *Lobby) GetGameServers() []GameServer {

	return lobby.Gamelist

}
