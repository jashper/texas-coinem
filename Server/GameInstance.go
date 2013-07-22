package Server

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
)

type GameLogic interface {
	UpdateState(playerID int, command string, game *GameInstance)
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
	isPlayerAllIn        []bool

	deck    [52]int
	deckIdx int

	handId    int
	blindsId  int
	sb        float64
	bb        float64
	ante      float64
	amtToCall float64
	prevBet   float64

	playerPots   []float64
	playerStacks []float64
	legalActions LegalActions
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
	this.isPlayerAllIn = make([]bool, parameters.PlayerCount)

	this.handId = -1
	this.blindsId = 0

	this.playerPots = make([]float64, parameters.PlayerCount)
	this.playerStacks = make([]float64, parameters.PlayerCount)
	this.legalActions = LegalActions{}
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
		check := this.legalActions.check

		if check {
			command = "CHECK"
		} else {
			command = "FOLD"
		}
		/// TODO: Send local "sitting-out" message
	}

	this.logic.UpdateState(int(playerID), command, this)
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
		check := this.legalActions.check

		if check {
			this.TakeTurn(playerID, "CHECK", false, 0)
		} else {
			this.TakeTurn(playerID, "FOLD", false, 0)
		}
	}
}

func (this *GameInstance) newHand() {
	select {
	case interrupt := <-this.interrupts:
		this.handleInterrupt(interrupt)
	default:
		break
	}

	this.deck = <-this.context.Entropy.Decks
	this.deckIdx = 0
	this.handId++
	this.boardCards = this.boardCards[0:0]

	this.playerQueue = this.playerQueue[0:0]
	for p := 0; p < this.parameters.PlayerCount; p++ {
		this.playerCards[p] = this.playerCards[p][0:0]
		this.playerPots[p] = 0
		this.isPlayerAllIn[p] = false

		if this.playerStacks[p] > 0 {
			this.playerQueue = append(this.playerQueue, p)
		}
	}

	if len(this.playerQueue) == 1 {
		// TODO: Signal end game
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

	for i := 0; i < len(this.playerQueue); i++ {
		if this.playerQueue[i] == this.buttonPlayer {
			this.playerQueueActiveIdx = i
			break
		}
	}

	// TODO: Is this valid for heads-up (ie: the blinds and actionPlayer)?

	sbPlayer := this.getNextPlayer()
	sbChips := this.getAvailableChips(sbPlayer)
	if this.sb > sbChips {
		this.playerPots[sbPlayer] += sbChips
		this.setAllIn(sbPlayer)
	} else {
		this.playerPots[sbPlayer] += this.sb
	}

	bbPlayer := this.getNextPlayer()
	bbChips := this.getAvailableChips(bbPlayer)
	if this.bb > bbChips {
		this.playerPots[bbPlayer] += bbChips
		this.setAllIn(bbPlayer)
	} else {
		this.playerPots[bbPlayer] += this.bb
	}

	if this.ante > 0 {
		for i := 0; i < len(this.playerQueue); i++ {
			p := this.playerQueue[i]
			chipsAvailable := this.getAvailableChips(p)
			if this.ante > chipsAvailable {
				this.playerPots[p] += chipsAvailable
				this.setAllIn(p)
			} else {
				this.playerPots[p] += this.ante
			}
		}
	}

	this.actionPlayer = -1
	this.amtToCall = this.bb
	this.prevBet = this.amtToCall

	this.logic.DealCards(this)

	this.broadcastMessage("Start hand #" +
		strconv.FormatInt(int64(this.handId), 10))
	this.broadcastMessage("Button: " + this.getUsername(this.buttonPlayer) + " | " +
		"SB: " + this.getUsername(sbPlayer) + " | " + "BB: " + this.getUsername(bbPlayer))

	this.newTurn(int32(this.getNextPlayer()))
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
	this.legalActions = LegalActions{}
	la := &this.legalActions

	minLimit := this.getMinLimit(playerID)
	maxLimit := this.getMaxLimit(playerID)

	unPaid := this.amtToCall - this.playerPots[playerID]
	chips := this.getAvailableChips(playerID)

	if unPaid == 0 {
		la.check = true
		if chips >= this.bb {
			la.bet = true
			la.min = this.bb
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

	fmt.Println(this.legalActions)

}

func (this *GameInstance) parseRequestedAction(command string) (action GameAction, value float64) {
	la := this.legalActions

	isCheckDefault := la.check

	split := strings.Split(command, ":")

	if split[0] == "CALL" && la.call {
		action = CALL
		return
	} else if split[0] == "BET" && la.bet {
		if len(split) == 2 {
			tempValue, _ := strconv.ParseFloat(split[1], 64)

			if la.min <= tempValue && tempValue <= la.max {
				action = BET
				value = tempValue
				return
			}
		}
	} else if split[0] == "RAISE" && la.raise {
		if len(split) == 2 {
			tempValue, _ := strconv.ParseFloat(split[1], 64)

			if la.min <= tempValue && tempValue <= la.max {
				action = RAISE
				value = tempValue
				return
			}
		}
	} else if split[0] == "ALLIN" && la.allin {
		action = ALLIN
		return
	}

	if isCheckDefault {
		action = CHECK
	} else {
		action = FOLD
	}

	return
}

func (this *GameInstance) getMinLimit(playerID int) float64 {
	unPaid := this.amtToCall - this.playerPots[playerID]

	if this.parameters.Limit == NO_LIMIT {
		if this.prevBet > (this.getAvailableChips(playerID) - unPaid) {
			return 0
		} else {
			return this.prevBet
		}
	}

	return -1
}

func (this *GameInstance) getMaxLimit(playerID int) float64 {
	unPaid := this.amtToCall - this.playerPots[playerID]

	if this.parameters.Limit == NO_LIMIT {
		return (this.getAvailableChips(playerID) - unPaid)
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

	// accounting for newHand button increment
	firstPlayer--
	if firstPlayer == -1 {
		firstPlayer = this.parameters.PlayerCount - 1
	}

	return firstPlayer
}

func (this *GameInstance) setAllIn(playerID int) {
	this.isPlayerAllIn[playerID] = true
	this.removeFromQueue(playerID)
}

func (this *GameInstance) removeFromQueue(playerID int) {
	pq := this.playerQueue
	for i := 0; i < len(pq); i++ {
		if pq[i] == playerID {
			pq = append(pq[:i], pq[i+1:]...)
		}
	}
	this.playerQueueActiveIdx--
}

func (this *GameInstance) getNewCard() int {
	card := this.deck[this.deckIdx]
	this.deckIdx++
	return card
}

func (this *GameInstance) getUsername(playerID int) string {
	return this.connections[playerID].userName
}

func (this *GameInstance) broadcastMessage(message string) {
	for i := 0; i < len(this.connections); i++ {
		c := this.connections[i]
		err := c.Write(c.userName + "-> " + message + "\n")
		if err != nil {
			this.isPlayerActive[i] = false
		}
	}
}

func (this *GameInstance) endHand() {
	//TODO: Send out summary of winners/losers

	for len(this.boardCards) < 5 {
		this.boardCards = append(this.boardCards,
			this.getNewCard())
	}

	inThePot := this.playerQueue
	for i := 0; i < len(this.isPlayerAllIn); i++ {
		if this.isPlayerAllIn[i] {
			inThePot = append(inThePot, i)
		}
	}

	inQueue := make([]bool, this.parameters.PlayerCount)
	for i := 0; i < len(inThePot); i++ {
		inQueue[inThePot[i]] = true
	}
	foldedPot := make([]int, 0)
	for i := 0; i < this.parameters.PlayerCount; i++ {
		if this.playerPots[i] != 0 && !inQueue[i] {
			foldedPot = append(foldedPot, i)
		}
	}
	inThePot = append(inThePot, foldedPot...)

	pots := make([]float64, len(inThePot)-len(foldedPot))
	strengths := make([]int, len(inThePot)-len(foldedPot))
	for i := 0; i < len(inThePot)-len(foldedPot); i++ {
		p := inThePot[i]

		pot := this.playerPots[p]
		pots[i] = pot

		var cards [7]uint
		for j := 0; j < 5; j++ {
			cards[j] = uint(this.boardCards[j])
		}
		for j := 5; j < 7; j++ {
			cards[j] = uint(this.playerCards[p][j-5])
		}

		strength, _ := this.context.HandEval.HandInfo(cards)
		strengths[i] = int(strength)
	}

	unsortedStrengths := make([]int, len(strengths))
	copy(unsortedStrengths, strengths)
	sort.Ints(strengths) // increasing order
	type payoutList struct {
		players []int
		next    *payoutList
	}
	toPay := payoutList{make([]int, 0), nil}
	tempPtr := &toPay
	for i := len(strengths) - 1; i >= 0; i-- {
		strength := strengths[i]
		count := 1
		if i-1 >= 0 {
			for strength == strengths[i-1] {
				count++
			}
			i -= count - 1
		}

		for j := 0; j < len(inThePot); j++ {
			if count == 0 {
				break
			}
			if unsortedStrengths[j] == strength {
				tempPtr.players = append(tempPtr.players, j)
				count--
			}
		}

		if i-1 >= 0 {
			tempPtr.next = &payoutList{make([]int, 0), nil}
			tempPtr = tempPtr.next
		}
	}

	for a := 0; a < len(inThePot); a++ {
		p := inThePot[a]
		pot := this.playerPots[p]
		tempPtr = &toPay

		changeToStack := float64(0)
		for pot > 0 {
			playersToPay := tempPtr.players

			donePaying := false
			for b := 0; b < len(playersToPay); b++ {
				if playersToPay[b] == a {
					donePaying = true
					break
				}
			}
			if donePaying {
				break
			}

			sortedPots := make([]float64, len(playersToPay))
			for b := 0; b < len(playersToPay); b++ {
				sortedPots[b] = pots[playersToPay[b]]
			}
			sort.Float64s(sortedPots)

			idxOrder := make([]int, len(playersToPay))
			hasBeenSeen := make([]bool, len(playersToPay))
			for i := 0; i < len(sortedPots); i++ {
				for j := 0; j < len(playersToPay); j++ {
					if pots[playersToPay[j]] == sortedPots[i] && !hasBeenSeen[j] {
						idxOrder[i] = j
						break
					}
				}
			}

			toPay := make([]float64, len(playersToPay))
			for b := 0; b < len(sortedPots); b++ {
				toPayIdx := idxOrder[b]

				greaterThan := float64(len(sortedPots) - b)

				if b != 0 {
					toPay[toPayIdx] = toPay[idxOrder[b-1]]
					if sortedPots[b] == sortedPots[b-1] {
						continue
					}
					toPay[toPayIdx] += (sortedPots[b] - sortedPots[b-1]) / greaterThan
				} else {
					toPay[toPayIdx] = sortedPots[0] / greaterThan
				}
			}

			for b := 0; b < len(toPay); b++ {
				this.playerStacks[inThePot[playersToPay[b]]] += toPay[b]
				pot -= toPay[b]
				changeToStack -= toPay[b]
			}

			tempPtr = tempPtr.next
		}

		this.playerStacks[p] += changeToStack
	}

	this.newHand()
}
