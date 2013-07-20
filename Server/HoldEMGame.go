package Server

type HoldEMGame struct {
	GameLogic
}

func (this HoldEMGame) UpdateState(playerID int32, command string, game *GameInstance) {
	action, value := game.parseRequestedAction(command)

	unPaid := game.amtToCall - game.playerPots[playerID]
	if action == FOLD {
		game.removeFromQueue(playerID)
	} else if action == CHECK {
		// do nothing
	} else if action == CALL {
		game.playerPots[playerID] += unPaid
	} else if action == BET {
		game.actionPlayer = playerID
		game.amtToCall += value
		game.prevBet = value
		game.playerPots[playerID] += unPaid + value

		if game.getAvailableChips(playerID) == 0 {
			game.setAllIn(playerID)
		}
	} else if action == RAISE {
		game.actionPlayer = playerID
		game.amtToCall += value
		game.prevBet = value
		game.playerPots[playerID] += unPaid + value

		if game.getAvailableChips(playerID) == 0 {
			game.setAllIn(playerID)
		}
	} else if action == ALLIN {
		chips := game.getAvailableChips(playerID)

		if unPaid < chips {
			game.actionPlayer = playerID
			game.amtToCall += (chips - unPaid)
		}

		game.playerPots[playerID] += chips
		game.setAllIn(playerID)
	}
}

func (this HoldEMGame) DealCards(game *GameInstance) {
	for i := 0; i < len(game.playerQueue); i++ {
		p := game.playerQueue[i]
		game.playerCards[p] = append(game.playerCards[p],
			game.getNewCard(), game.getNewCard())
	}

	//TODO: send card pairs out to appropriate users
}
