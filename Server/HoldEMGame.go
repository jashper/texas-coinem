package Server

type HoldEMGame struct {
	GameLogic
}

func (this HoldEMGame) UpdateState(playerID int32, command string, game *GameInstance) {

}

func (this HoldEMGame) DealCards(game *GameInstance) {
	for i := 0; i < len(game.playerQueue); i++ {
		p := game.playerQueue[i]
		game.playerCards[p] = append(game.playerCards[p],
			game.getNewCard(), game.getNewCard())
	}

	//TODO: send card pairs out to appropriate users
}
