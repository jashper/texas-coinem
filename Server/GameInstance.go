package Server

import (
	"math"
	"sync/atomic"
)

/*
	This class is a state machine representing the mechanics for a single
	poker table. It is responsible for delegating turn order and communicating
	with players of any updates to the state of the game. Commands sent by players
	to the server are parsed and processed in a non-blocking manner, by using
	a single method, TakeTurn(), to give access to goprocesses.  This method
	utilizes an atomic compare and swap to limit the usage of this class to only
	one goprocess at a time (ie: the appropriate player or timer for that player)
*/

// ############################################
//     Helper Structs
// ############################################

// Different game variants implement this interface (ie: HoldEM, Stud, etc.)
type GameLogic interface {
	UpdateState(playerID int, action GameAction, game *GameInstance)
	DealCards(game *GameInstance)
}

// Used to keep track of player game-state info
type PlayerInfo struct {
	isConnected bool // is currently connected to the server
	stack       float64
	pot         float64
	hand        []int
	hasFolded   bool
}

// ############################################
//     Constructor Struct & Init
// ############################################

type GameInstance struct {
	context    *ServerContext
	parameters GameParameters
	logic      GameLogic

	connections []*Connection
	interrupts  chan GameInterrupt
	timerID     int // used to guarantee that a "stale" timer
	// can't take a turn that has already been taken (ie: this
	// value is incremented for each new timer goprocess spawn)

	players       []PlayerInfo
	buttonPlayer  int
	currentPlayer int32 // its purpose is to guarantee that only
	// one goprocess can take a turn and update the state of the game

	currentLA       LegalActions
	streetEndPlayer int // denotes the player that will end the
	// current street

	activeCount int // how many players can still take a turn
	// during a hand (ie: how many haven't folded, aren't all-in,
	// or aren't out of the game)

	deck       [52]int
	deckIdx    int
	boardCards []int

	handID     int
	blindLevel int
	sb         float64
	bb         float64
	ante       float64
	maxPot     float64 // largest amount a single player has
	// committed to the pot

	minBet float64 // minimum bet or raise a player must make
	// if he wishes to bet or raise
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
	g.interrupts = make(chan GameInterrupt, 25) // TODO: Big enough buffer?
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

	g.handID = -1 // make sure nothing can take a turn until newHand()
	// is run for the first time

	g.blindLevel = 0

	go QueueBlindsTimer(parameters.LevelTime, g)
	g.newHand()
}

// ############################################
//     Turn taking & game state update methods
// ############################################

// The only method that gets called outside of GameInstance; it's
// responsible for updating the state of the game after each
// player tries to take an action
func (g *GameInstance) TakeTurn(playerID int, action GameAction,
	isTimer bool, timerID int) {

	// check against "stale" timers
	if isTimer && timerID != g.timerID {
		return
	}

	// non-blocking way of guaranteeing that only the right player or timer
	// for that player can take a turn and update the state
	if !atomic.CompareAndSwapInt32(&g.currentPlayer, int32(playerID), -1) {
		return
	}

	if isTimer {
		g.players[playerID].isConnected = false // player is now sitting out
		// TODO: Send local "sitting-out" message
	}

	// TODO: make sure this action is legal

	g.logic.UpdateState(playerID, action, g)
}

// Prepares the game-state for the next player to take a turn
func (g *GameInstance) newTurn(playerID int) {
	g.updateLegalActions(playerID)
	g.currentPlayer = int32(playerID)

	//TODO: Send out new turn message

	isConnected := g.players[playerID].isConnected
	if isConnected { // set a timer for the player to take a turn
		g.timerID++
		go QueueTurnTimer(playerID, g.timerID, g.parameters.TurnTime, g)
	} else { // if sitting out, automatically take his turn
		action := GameAction{}
		canCheck := g.currentLA.check
		if canCheck {
			action.aType = CHECK
		} else {
			action.aType = FOLD
		}

		g.TakeTurn(playerID, action, false, 0)
	}

}

// Prepares the game-state for a new hand to start
func (g *GameInstance) newHand() {
	// process server interrupts (useful for tournaments & rebalancing tables)
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

	g.activeCount = 0
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]
		p.hand = p.hand[0:0]
		p.pot = 0
		p.hasFolded = false

		if p.stack > 0 {
			g.activeCount++
		}
	}

	if g.activeCount == 1 {
		//TODO: Signal end of game
	}

	level := g.parameters.Blinds.levels[g.blindLevel]
	g.sb = level.sb
	g.bb = g.sb * 2
	g.ante = level.ante

	g.buttonPlayer = g.getNextPlayer(g.buttonPlayer)

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

	g.streetEndPlayer = -1 // needed for the odd nature of pre-flop, where
	// blinds are in place instead of an initial bet; whenever a player bets
	// or raises, this value will change to their playerID
	g.maxPot = g.bb + g.ante
	g.minBet = g.bb

	g.logic.DealCards(g)
	g.newTurn(g.getNextPlayer(bbIdx))
}

// Deals additional cards if players went all-in before the last street, as
// well as determines hand strengths and the amount of chips each player
// either gains or loses
func (g *GameInstance) endHand(handSize, boardSize int) {
	// deal additional cards if necessary
	for len(g.boardCards) < boardSize {
		g.boardCards = append(g.boardCards,
			g.getNewCard())
	}

	// compute player hand strengths relative to the board
	pots := make([]float64, len(g.players))
	strengths := make([]int, len(g.players))
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]

		pots[i] = p.pot

		var cards [7]uint
		for j := 0; j < boardSize; j++ {
			cards[j] = uint(g.boardCards[j])
		}
		for j := boardSize; j < 7; j++ {
			cards[j] = uint(p.hand[j-boardSize])
		}

		strength, _ := g.context.HandEval.HandInfo(cards)
		strengths[i] = int(strength)
	}

	// for each player, determine if that players owes chips to any
	// other players (ie: if they don't tie for or have the strongest hand)
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]

		toPay := make([]int, 0)
		toPayPots := make([]float64, 0)
		toPayStrengths := make([]int, 0)

		// find all players with better ranked hands
		strength := strengths[i]
		for j := 0; j < len(g.players); j++ {
			if strengths[j] > strength {
				toPay = append(toPay, j)
				toPayPots = append(toPayPots, pots[j])
				toPayStrengths = append(toPayStrengths, strengths[j])
			}
		}

		if len(toPay) == 0 {
			continue
		}

		// find the best of the players that are better than you;
		// there are multiple if they tie with each other
		maxI := []int{0}
		maxStrength := toPayStrengths[0]
		for k := 1; k < len(toPayStrengths); k++ {
			if toPayStrengths[k] > maxStrength {
				maxI = []int{k}
				maxStrength = toPayStrengths[k]
			} else if toPayStrengths[k] == maxStrength {
				maxI = append(maxI, k)
			}
		}

		// complicated logic to determine how many chips are owed
		// to each player; accounts for infinite side-pots; have fun
		// trying to figure this part out
		prevMin := float64(-1)
		for len(maxI) > 0 {
			minI := []int{0}
			minPot := toPayPots[maxI[0]]
			for k := 1; k < len(maxI); k++ {
				if toPayPots[maxI[k]] < minPot {
					minI = []int{k}
					minPot = toPayPots[maxI[k]]
				} else if toPayPots[maxI[k]] == minPot {
					minI = append(minI, k)
				}
			}
			payMult := len(maxI)
			for k := 0; k < len(minI); k++ {
				idx := minI[k]
				minI[k] = maxI[idx]
				maxI = append(maxI[:idx], maxI[idx+1:]...)
			}

			var amtToPay float64
			if prevMin == -1 {
				amtToPay = math.Min(float64(payMult)*minPot, p.pot)
			} else {
				amtToPay = math.Min(float64(payMult)*(minPot-prevMin), p.pot)
			}

			p.pot -= amtToPay
			p.stack -= amtToPay
			amtToPay /= float64(payMult)

			for k := 0; k < len(minI); k++ {
				g.players[toPay[minI[k]]].stack += amtToPay
			}
			for k := 0; k < len(maxI); k++ {
				g.players[toPay[maxI[k]]].stack += amtToPay
			}

			prevMin = minPot
		}
	}
}

// ############################################
//     Helper methods
// ############################################

// Used to randomly determine the first person to have the button
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
			maxCard = g.deck[i]
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

// Determine the valid actions a player can take, before calling
// newTurn() and broadcasting the fact that it's his turn
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
			la.allin = true // incomplete bet, type of all-in
		}
	} else {
		la.fold = true
		if unPaid >= chips {
			la.allin = true // can't cover previous bet or raise, type of all-in
		} else {
			la.call = true
			if minLimit == 0 {
				la.allin = true
			} else {
				la.raise = true // incomplete raise, type of all-in
				la.min = minLimit
				la.max = maxLimit
			}
		}
	}
}

// Determine the minimum amount a player must bet or raise by, depending
// on the type of game
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

// Determine the maximum amount a player can bet or raise by, depending
// on the type of game
func (g *GameInstance) getMaxLimit(playerID int) (limit float64) {
	unPaid := g.getChipsUnpaid(playerID)

	if g.parameters.Limit == NO_LIMIT {
		limit = (g.getChipsAvailable(playerID) - unPaid)
	}

	return
}

// Find out how many more chips a player can put in the pot
func (g *GameInstance) getChipsAvailable(playerID int) float64 {
	return (g.players[playerID].stack - g.players[playerID].pot)
}

// Find out how many chips the player needs to put in the pot to call
func (g *GameInstance) getChipsUnpaid(playerID int) float64 {
	return (g.maxPot - g.players[playerID].pot)
}

// Execute the appropriate logic for various types of interrupts
func (g *GameInstance) handleInterrupt(interrupt GameInterrupt) {
	// Increment the blinds level and queue the timer for the next level
	if interrupt.iType == I_GAME_NEW_BLINDS {
		g.blindLevel++
		go QueueBlindsTimer(g.parameters.LevelTime, g)
	}
}

// Find the next active player in the game by looping rightwards/clockwise
// through the player list; this player hasn't folded, isn't all-in,
// and is still in the game (ie: non-zero stack)
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

// Grab the next card in the deck for this turn
func (g *GameInstance) getNewCard() int {
	card := g.deck[g.deckIdx]
	g.deckIdx++
	return card
}
