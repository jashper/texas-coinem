package Server

import (
	"fmt"
	"sync/atomic"
)

type GameLogic interface {
	UpdateState(playerID int32, command string, game *GameInstance)
	DealCards(game *GameInstance)
}

type GameInstance struct {
	context     *ServerContext
	connections []*Connection
	parameters  GameParameters
	interrupts  chan GameInterrupt
	logic       GameLogic

	activePlayer   int32
	timerID        int
	isPlayerActive []bool

	playerQueue          []int
	playerQueueActiveIdx int
	actionPlayer         int // most recent player to initiate action
	buttonPlayer         int

	handId    int
	blindsId  int
	sb        float64
	bb        float64
	ante      float64
	amtToCall float64
	prevBet   float64

	playerPots   []float64
	playerStacks []float64
	legalActions []string
	playerCards  [][]int
	boardCards   []int
}

func (this *GameInstance) Init(context *ServerContext, connections []*Connection,
	parameters GameParameters) {

	this.context = context
	this.connections = connections
	this.parameters = parameters
	this.interrupts = make(chan GameInterrupt, 25)
	switch {
	case parameters.Variant == HOLDEM:
		this.logic = HoldEMGame{}
	}

	this.activePlayer = -1
	this.timerID = 0
	this.isPlayerActive = make([]bool, parameters.PlayerCount)

	this.playerQueue = make([]int, 0)

	this.handId = 0
	this.blindsId = 0

	this.playerPots = make([]float64, parameters.PlayerCount)
	this.playerStacks = make([]float64, parameters.PlayerCount)
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

	this.buttonPlayer = this.firstButtonPosition()

	go QueueBlindsTimer(parameters.TurnTime, this)
	this.newHand()
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
	this.updateLegalActions(int(playerID))

	this.activePlayer = playerID

	/// TODO: Send local "new-turn" message

	active := this.isPlayerActive[playerID]
	if active {
		this.timerID++
		go QueueTurnTimer(playerID, this.timerID, this.parameters.TurnTime, this)
	}

	if !active {
		check := false
		for l := 0; l < len(this.legalActions); l++ {
			if this.legalActions[l] == "CHECK" {
				check = true
				break
			}
		}

		if check {
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

	for {
		select {
		case interrupt := <-this.interrupts:
			this.handleInterrupt(interrupt)
		default:
			break
		}
	}

	this.sb = this.parameters.Blinds.GetSB(this.blindsId)
	this.bb = this.sb * 2
	this.ante = this.parameters.Blinds.GetAnte(this.blindsId)

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

	sbPlayer := this.getNextPlayer()
	sbChips := this.getAvailableChips(sbPlayer)
	if this.sb > sbChips {
		this.playerPots[sbPlayer] += sbChips
	} else {
		this.playerPots[sbPlayer] += this.sb
	}

	bbPlayer := this.getNextPlayer()
	bbChips := this.getAvailableChips(bbPlayer)
	if this.bb > bbChips {
		this.playerPots[bbPlayer] += bbChips
	} else {
		this.playerPots[bbPlayer] += this.bb
	}

	if this.ante > 0 {
		for p := 0; p < len(this.playerQueue); p++ {
			chipsAvailable := this.getAvailableChips(p)
			if this.ante > chipsAvailable {
				this.playerPots[p] += chipsAvailable
			} else {
				this.playerPots[p] += this.ante
			}
		}
	}

	this.actionPlayer = this.getNextPlayer()
	this.amtToCall = this.bb
	this.prevBet = this.amtToCall

	this.logic.DealCards(this)

	this.newTurn(int32(this.actionPlayer))
}

func (this *GameInstance) getNextPlayer() int {
	if this.playerQueueActiveIdx == len(this.playerQueue)-1 {
		this.playerQueueActiveIdx = 0
	} else {
		this.playerQueueActiveIdx++
	}

	return this.playerQueueActiveIdx
}

func (this *GameInstance) getAvailableChips(playerID int) float64 {
	return (this.playerStacks[playerID] - this.playerPots[playerID])
}

func (this *GameInstance) updateLegalActions(playerID int) {
	la := this.legalActions
	la = la[0:0]

	minLimit := this.getMinLimit(playerID)
	maxLimit := this.getMaxLimit(playerID)

	if this.amtToCall == 0 {
		la = append(la, "CHECK")
		la = append(la, "BET"+fmt.Sprint(this.bb)+":"+fmt.Sprint(maxLimit))
	} else {
		la = append(la, "CALL")
		la = append(la, "FOLD")
		if minLimit == 0 {
			la = append(la, "ALLIN")
		} else {
			la = append(la, "RAISE"+fmt.Sprint(minLimit)+":"+fmt.Sprint(maxLimit))
		}
	}

}

func (this *GameInstance) getMinLimit(playerID int) float64 {
	if this.parameters.Limit == NO_LIMIT {
		if this.prevBet >= this.getAvailableChips(playerID) {
			return 0
		} else {
			return this.prevBet
		}
	}

	return -1
}

func (this *GameInstance) getMaxLimit(playerID int) float64 {
	if this.parameters.Limit == NO_LIMIT {
		return this.getAvailableChips(playerID)
	}

	return -1
}

func (this *GameInstance) handleInterrupt(interrupt GameInterrupt) {
	if interrupt.iType == I_GAME_NEW_BLINDS {
		this.blindsId++
		go QueueBlindsTimer(this.parameters.TurnTime, this)
	}
}

func (this *GameInstance) firstButtonPosition() int {
	deck := <-this.context.Entropy.Decks

	firstPlayer := 0
	maxCard := deck[0]
	for i := 1; i < this.parameters.PlayerCount; i++ {
		if deck[i] > maxCard {
			firstPlayer = i
			maxCard = deck[i]
		}
	}

	return firstPlayer
}
