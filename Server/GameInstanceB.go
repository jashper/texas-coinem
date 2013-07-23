package Server

import (
	"sync/atomic"
)

type GameLogic interface {
	UpdateState(playerID int, command string, game *GameInstance)
	DealCards(game *GameInstance)
}

type PlayerInfo struct {
	isConnected bool
	stack       float64
	pot         float64
	hand        []int
}

type GameInstance struct {
	context    *ServerContext
	parameters GameParameters
	logic      GameLogic

	connections []*Connection
	interrupts  chan GameInterrupt
	timerID     int

	players         []PlayerInfo
	buttonPlayer    int
	currentPlayer   int
	currentLA       LegalActions
	streetEndPlayer int

	deck       [52]int
	deckIdx    int
	boardCards []int

	handID     int
	blindLevel int
	sb         float64
	bb         float64
	ante       float64
	maxPot     float64
	minBet     float64
}

func (g *GameInstance) Init(context *ServerContext, parameters GameParameters,
	connections []*Connection) {

	g.context = context
	g.parameters = parameters
	switch {
	case parameters.Variant == HOLDEM:
		g.logic = HoldEMGame{}
	}

	g.connections = connections
	g.interrupts = make(chan GameInterrupt, 25) // Big enough buffer?
	g.timerID = 0

	g.players = make([]PlayerInfo, parameters.PlayerCount)
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]
		p.isConnected = true
		p.stack = parameters.ChipCount
		p.hand = make([]int, 0)
	}
	g.buttonPlayer = g.firstButtonPosition()
	g.currentPlayer = -1
	g.currentLA = LegalActions{}

	g.boardCards = make([]int, 0)

	g.handID = -1
	g.blindLevel = 0

	go QueueBlindsTimer(parameters.LevelTime, g)
	g.newHand()
}

func (g *GameInstance) TakeTurn(playerID int, action GameAction,
	isTimer bool, timerID int) {

	if isTimer && timerID != g.timerID {
		return
	}

	if !atomic.CompareAndSwapInt32(&g.currentPlayer, int32(playerID), -1) {
		return
	}

	if isTimer {
		g.players[playerID].isConnected = false
		// TODO: Send local "sitting-out" message
	}

	this.logic.UpdateState(playerID, action, this)
}

func (g *GameInstance) newTurn(playerID int32) {
	g.updateLegalActions(playerID)
	g.currentPlayer = playerID

	//TODO: Send out new turn message

	isConnected := g.players[playerID].isConnected
	if isConnected {
		g.timerID++
		go QueueTurnTimer(playerID, g.timerID, g.parameters.TurnTime, g)
	} else {
		action := GameAction{}
		canCheck := g.currentLA.check
		if canCheck {
			action.aType = CHECK
		} else {
			action.aType = FOLD
		}

		game.TakeTurn(playerID, action, false, 0)
	}

}

func (g *GameInstance) newHand() {

}

// ######################
//       Helpers
// ######################

func (g *GameInstance) firstButtonPosition() int {
	tempDeck := <-g.context.Entropy.Decks

	// TODO: Fix this logic to draw a new deck
	//       for ties between th same type of card
	//       (ie: take mod of maxCard)

	firstPlayer := 0
	maxCard := tempDeck[0]
	for i := 1; i < len(g.players); i++ {
		if tempDeck[i] > maxCard {
			firstPlayer = i
			maxCard = deck[i]
		}
	}

	// TODO: notify players of who drew what
	//       and who has the button

	// accounts for the button increment to come in newHand()
	firstPlayer--
	if firstPlayer == -1 {
		firstPlayer = len(g.players) - 1
	}

	return firstPlayer
}

func (g *GameInstance) updateLegalActions(playerID int) {
	g.currentLA = LegalActions{}
	la := &g.currentLA

	minLimit := g.getMinLimit(playerID)
	maxLimit := g.getMaxLimit(playerID)

	unPaid := g.getChipsUnpaid(playerID)
	chips := g.getChipsAvailable(playerID)

	if unPaid == 0 {
		la.check = true
		if chips >= g.bb {
			la.bet = true
			la.min = g.bb
			la.max = maxLimit
		} else {
			la.allin = true
		}
	} else {
		la.fold = true
		if unPaid >= chips {
			la.allin = true
		} else {
			la.call = true
			if minLimit == 0 {
				la.allin = true
			} else {
				la.raise = true
				la.min = minLimit
				la.max = maxLimit
			}
		}
	}
}

func (g *GameInstance) getMinLimit(playerID int) (limit float64) {
	unPaid := g.getChipsUnpaid(playerID)

	if g.parameters.Limit == NO_LIMIT {
		if g.minBet > (g.getChipsAvailable(playerID) - unPaid) {
			limit = 0
		} else {
			limit = g.minBet
		}
	}

	return
}

func (this *GameInstance) getMaxLimit(playerID int) (limit float64) {
	unPaid := g.getChipsUnpaid(playerID)

	if g.parameters.Limit == NO_LIMIT {
		limit = (g.getChipsAvailable(playerID) - unPaid)
	}

	return
}

func (g *GameInstance) getChipsAvailable(playerID int) float64 {
	return (g.players[playerID].stack - g.players[playerID].pot)
}

func (g *GameInstance) getChipsUnpaid(playerID int) float64 {
	return (g.maxPot - g.players[playerID].pot)
}
