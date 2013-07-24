package Server

/*
	This class implements a HoldEm version of GameLogic to be used
	by GameInstance.

	Side-Note: Most of this logic will probably overlap with other
	variants of GameLogic, and can probably be moved to GameInstance
*/

type HoldEMGame struct {
	GameLogic
}

// ############################################
//     Game state update methods
// ############################################

// Processes a player/timer's requested action relative to the
// current state of the game
func (this HoldEMGame) UpdateState(playerID int, action GameAction, g *GameInstance) {
	p := g.players[playerID]

	unPaid := g.getChipsUnpaid(playerID)
	if action.aType == FOLD {
		p.hasFolded = true
		g.activeCount-- // out for the rest of the turn
	} else if action.aType == CHECK {
		//do nothing
	} else if action.aType == BET || action.aType == RAISE {
		g.streetEndPlayer = playerID // this player now marks the end of a street
		g.maxPot += action.value
		g.minBet = action.value
		p.pot += unPaid + action.value // unPaid should be 0 for a BET

		if g.getChipsAvailable(playerID) == 0 { // is player all-in?
			g.activeCount--
		}
	} else if action.aType == ALLIN {
		chips := g.getChipsAvailable(playerID)

		if unPaid < chips { // check if this is an incomplete bet/raise type of all-in
			// instead of a call type of all-in
			g.streetEndPlayer = playerID
			g.maxPot += chips - unPaid
		}

		p.pot += chips
		g.activeCount--
	}

	nextPlayer := g.getNextPlayer(playerID)
	endOfStreet := false

	if len(g.boardCards) == 0 { // necessary because streetEndPlayer will be -1 for preFlop
		if action.aType == CHECK || g.activeCount == 1 {
			endOfStreet = true
		}
	}
	if g.streetEndPlayer == nextPlayer {
		endOfStreet = true
	}

	if !endOfStreet {
		g.newTurn(nextPlayer)
	} else {
		if g.activeCount == 1 || len(g.boardCards) == 5 { // is only one active player left
			// or is the river done yet
			g.endHand(2, 5)
		} else {
			g.streetEndPlayer = g.getNextPlayer(g.buttonPlayer) // the closest player to the button
			// always starts the street after the preflop
			g.minBet = 0

			g.boardCards = append(g.boardCards, g.getNewCard())
			if len(g.boardCards) == 1 { // draw more than one for the flop
				g.boardCards = append(g.boardCards,
					g.getNewCard(), g.getNewCard())
			}

			g.newTurn(g.streetEndPlayer)
		}
	}
}

// Deal/draw the initial hands for each player
func (this HoldEMGame) DealCards(g *GameInstance) {
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]
		p.hand = append(p.hand, g.getNewCard(), g.getNewCard())
	}
}
