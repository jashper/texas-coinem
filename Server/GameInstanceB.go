package Server

import (
	"math"
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
	hasFolded   bool
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
	activeCount     int

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
	select {
	case interrupt := <-g.interrupts:
		g.handleInterrupt(interrupt)
	default:
		break
	}

	g.deck = <-g.context.Entropy.Decks
	g.deckIdx = 0
	g.handID++
	g.boardCards = g.boardCards[0:0]

	activeCount = 0
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]
		p.hand = p.hand[0:0]
		p.pot = 0
		p.hasFolded = false

		if p.stack > 0 {
			activeCount++
		}
	}

	if activeCount == 1 {
		//TODO: Signal end of game
	}

	level := g.parameters.Blinds.levels[g.blindLevel]
	g.sb = level.sb
	g.bb = g.sb * 2
	g.ante = level.ante

	g.buttonPlayer == g.getNextPlayer(g.buttonPlayer)

	sbIdx := g.getNextPlayer(g.buttonPlayer)
	sbPlayer := g.players[sbIdx]
	if g.sb > sbPlayer.stack {
		sbPlayer.pot = sbPlayer.stack
	} else {
		sbPlayer.pot = g.sb
	}

	bbIdx := g.getNextPlayer(sbIdx)
	bbPlayer := g.players[bbIdx]
	if g.bb > bbPlayer.stack {
		bbPlayer.pot = bbPlayer.stack
	} else {
		bbPlayer.pot = g.bb
	}

	if g.ante > 0 {
		for i := 0; i < len(g.players); i++ {
			chips := g.getChipsAvailable(i)
			if g.ante > chips {
				g.players[i].pot += chips
			} else {
				g.players[i].pot += g.ante
			}
		}
	}

	g.streetEndPlayer = -1
	g.maxPot = g.bb + g.ante
	g.minBet = g.bb

	g.logic.DealCards(g)
	g.newTurn(g.getNextPlayer(bbIdx))
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

func (g *GameInstance) handleInterrupt(interrupt GameInterrupt) {
	if interrupt.iType == I_GAME_NEW_BLINDS {
		g.blindLevel++
		go QueueBlindsTimer(g.parameters.LevelTime, g)
	}
}

func (g *GameInstance) getNextPlayer(current int) int {
	for g.players[current].pot == g.players[current].stack ||
		g.players[current].hasFolded {
		if current < len(g.players)-1 {
			current++
		} else {
			current = 0
		}
	}

	return current
}
