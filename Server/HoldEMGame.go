package Server

type HoldEMGame struct {
	GameLogic
}

func (this HoldEMGame) UpdateState(playerID int, action GameAction, g *GameInstance) {
	p := g.players[playerID]

	unPaid := g.getChipsUnpaid(playerID)
	if action.aType == FOLD {
		p.hasFolded = true
		g.activeCount--
	} else if action.aType == CHECK {
		//do nothing
	} else if action.aType == BET || action.aType == RAISE {
		g.streetEndPlayer = playerID
		g.maxPot += action.value
		g.minBet = action.value
		p.pot += unPaid + action.value

		if g.getChipsAvailable(playerID) == 0 {
			g.activeCount--
		}
	} else if action.aType == ALLIN {
		chips := g.getChipsAvailable(playerID)

		if unPaid < chips {
			g.streetEndPlayer = playerID
			g.maxPot += chips - unPaid
		}

		p.pot += chips
		g.activeCount--
	}

	nextPlayer := g.getNextPlayer(playerID)
	endOfStreet := false

	if len(g.boardCards) == 0 {
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
		if g.activeCount == 1 || len(g.boardCards) == 5 {
			g.endHand(2, 5)
		} else {
			g.streetEndPlayer = g.getNextPlayer(g.buttonPlayer)
			g.minBet = 0

			g.boardCards = append(g.boardCards, g.getNewCard())
			if len(g.boardCards) == 1 {
				g.boardCards = append(g.boardCards,
					g.getNewCard(), g.getNewCard())
			}

			g.newTurn(g.streetEndPlayer)
		}
	}
}

func (this HoldEMGame) DealCards(g *GameInstance) {
	for i := 0; i < len(g.players); i++ {
		p := g.players[i]
		p.hand = append(p.hand, g.getNewCard(), g.getNewCard())
	}
}
