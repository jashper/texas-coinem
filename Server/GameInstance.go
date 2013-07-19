package Server

import (
	"sync/atomic"
)

type GameLogic interface {
	UpdateState(playerID int32, command string, game *GameInstance)
}

type GameInstance struct {
	context     *ServerContext
	connections []*Connection
	parameters  GameParameters
	logic       GameLogic

	activePlayer   int32
	timerID        int
	isPlayerActive []bool

	playerQueue          []int
	playerQueueActiveIdx int
	actionPlayer         int // most recent player to initiate action
	buttonPlayer         int

	handId    int
	sb        int
	bb        int
	ante      int
	amtToCall int
	prevBet   int

	playerPots   []int
	playerStacks []int
	legalActions []string
	playerCards  [][]int
	boardCards   []int
}

func (this *GameInstance) Init(context *ServerContext, connections []*Connection,
	parameters GameParameters) {

	this.context = context
	this.connections = connections
	this.parameters = parameters
	switch {
	case parameters.Variant == HOLDEM:
		this.logic = HoldEMGame{}
	}

	this.activePlayer = -1
	this.timerID = 0
	this.isPlayerActive = make([]bool, parameters.PlayerCount)

	this.playerQueue = make([]int, 0)

	this.handId = 0

	this.playerPots = make([]int, parameters.PlayerCount)
	this.playerStacks = make([]int, parameters.PlayerCount)
	this.legalActions = make([]string, 0)
	this.playerCards = make([][]int, parameters.PlayerCount)
	for p := 0; p < parameters.PlayerCount; p++ {
		this.playerCards[p] = make([]int, 0)
	}
	this.boardCards = make([]int, 0)

	for i := 0; i < parameters.PlayerCount; i++ {
		this.isPlayerActive[i] = true
		this.playerStacks[i] = parameters.ChipCount
	}

	this.buttonPlayer = 0 // TODO: randomly pick this
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

	this.logic.UpdateState(playerID, command, this)
}

func (this *GameInstance) newTurn(playerID int32) {
	// TODO: updateLegalActions

	active := this.isPlayerActive[playerID]
	if active {
		this.timerID++
		go QueueHandTimer(playerID, this.timerID, this.parameters.TurnTime, this)
	}

	this.activePlayer = playerID

	/// TODO: Send local "new-turn" message

	if !active {
		checkPresent := false
		for l := 0; l < len(this.legalActions); l++ {
			if this.legalActions[l] == "CHECK" {
				checkPresent = true
				break
			}
		}

		if checkPresent {
			this.TakeTurn(playerID, "CHECK", false, 0)
		} else {
			this.TakeTurn(playerID, "FOLD", false, 0)
		}
	}
}

func (this *GameInstance) newHand() {
	// TODO: double check we clear/reset appropriate items

	this.handId++
	this.boardCards = this.boardCards[0:0]

	this.playerQueue = this.playerQueue[0:0]
	for p := 0; p < this.parameters.PlayerCount; p++ {
		this.playerCards[p] = this.playerCards[p][0:0]
		this.playerPots[p] = 0

		if this.playerStacks[p] > 0 {
			this.playerQueue = append(this.playerQueue, p)
		}
	}

	if len(this.playerQueue) == 1 {
		// TODO: Signal end game
	}

	// TODO: Handle Queue Overhead Interrupts
	// TODO: Initialize sb, bb, & ante and set blind timer

	if this.buttonPlayer < this.parameters.PlayerCount-1 {
		this.buttonPlayer++
	} else {
		this.buttonPlayer = 0
	}

	for this.playerStacks[this.buttonPlayer] == 0 {
		if this.buttonPlayer < this.parameters.PlayerCount-1 {
			this.buttonPlayer++
		} else {
			this.buttonPlayer = 0
		}
	}

	for p := 0; p < len(this.playerQueue); p++ {
		if this.playerQueue[p] == this.buttonPlayer {
			this.playerQueueActiveIdx = p
			break
		}
	}

	// TODO: Is this valid for heads-up (ie: the blinds and actionPlayer)?

}
