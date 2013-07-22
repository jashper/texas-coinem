package Server

import (
	"fmt"
)

type HoldEMGame struct {
	GameLogic
}

func (this HoldEMGame) UpdateState(playerID int, command string, game *GameInstance) {
	action, value := game.parseRequestedAction(command)

	unPaid := game.amtToCall - game.playerPots[playerID]
	if action == FOLD {
		game.removeFromQueue(playerID)
	} else if action == CHECK {
		// do nothing
	} else if action == CALL {
		game.playerPots[playerID] += unPaid
	} else if action == BET || action == RAISE {
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

	nextPlayer := game.getNextPlayer()

	endOfStreet := false

	if len(game.boardCards) == 0 {
		if action == CHECK && game.playerPots[playerID] == game.bb {
			endOfStreet = true
		} else if len(game.playerQueue) == 1 {
			endOfStreet = true
		}
	}

	if game.actionPlayer == nextPlayer {
		endOfStreet = true
	}

	actionMessage := game.getUsername(playerID) + "->" + command
	game.broadcastMessage(actionMessage)

	//TODO: send out summary of updates (ie:
	//   end of turn, new street, end of hand)

	if !endOfStreet {
		game.newTurn(int32(nextPlayer))
	} else {
		if len(game.playerQueue) == 1 || len(game.boardCards) == 5 {
			game.endHand()
		} else {
			for i := 0; i < len(game.playerQueue); i++ {
				if game.playerQueue[i] == game.buttonPlayer {
					game.playerQueueActiveIdx = i
					break
				}
			}
			game.actionPlayer = game.getNextPlayer()
			game.prevBet = 0

			game.boardCards = append(game.boardCards, game.getNewCard())
			if len(game.boardCards) == 1 {
				game.boardCards = append(game.boardCards,
					game.getNewCard(), game.getNewCard())
			}
			boardMessage := "New board: " + fmt.Sprint(game.boardCards)
			game.broadcastMessage(boardMessage)

			game.newTurn(int32(game.actionPlayer))
		}
	}

}

func (this HoldEMGame) DealCards(game *GameInstance) {
	for i := 0; i < len(game.playerQueue); i++ {
		p := game.playerQueue[i]
		game.playerCards[p] = append(game.playerCards[p],
			game.getNewCard(), game.getNewCard())

		message := "Hand: " + cardToString(game.playerCards[p][0]) + " " +
			cardToString(game.playerCards[p][1]) + "\n"
		game.connections[p].Write(message)
	}

	//TODO: send card pairs out to appropriate users
}
