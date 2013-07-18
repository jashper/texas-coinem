package Server

import (
	"sync/atomic"
)

type GameInstance struct {
	GameInstanceInterface

	context     *ServerContext
	connections []*Connection
	parameters  GameParameters

	activePlayer   int32
	timerID        int
	isPlayerActive []bool
}

func (this *GameInstance) Init(context *ServerContext, connections []*Connection, parameters GameParameters) {

	this.context = context
	this.connections = connections
	this.parameters = parameters

	this.activePlayer = -1
	this.timerID = 0
	this.isPlayerActive = make([]bool, parameters.PlayerCount)
	for i := 0; i < len(this.isPlayerActive); i++ {
		this.isPlayerActive[i] = true
	}
}

func (this *GameInstance) TakeTurn(playerID int32, command string,
	isTimer bool, timerID int) {

	if isTimer && timerID != this.timerID {
		return
	}

	if !atomic.CompareAndSwapInt32(&this.activePlayer, playerID, -1) {
		return
	}

	if isTimer {
		this.isPlayerActive[playerID] = false
		/// TODO: Send local "sitting-out" message
	}

	this.UpdateState(playerID, command)
}
