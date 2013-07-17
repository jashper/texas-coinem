package Server

import "net"

type GameInstance struct {
	Context     ServerContext
	Connections []net.Conn
	// TODO: GameParameters
}

func (this *GameInstance) TakeTurn(playerNum int, command string, isTimer bool) {

}
