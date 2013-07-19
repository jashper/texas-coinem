package Server

type HoldEMGame struct {
	GameLogic
}

func (this HoldEMGame) UpdateState(playerID int32, command string, game *GameInstance) {
}

func (this HoldEMGame) DealCards(game *GameInstance) {
}
