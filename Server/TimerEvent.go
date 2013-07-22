package Server

import (
	"time"
)

func QueueTurnTimer(playerID int32, timerID int, interval time.Duration, game *GameInstance) {
	time.Sleep(interval)

	game.TakeTurn(playerID, "", true, timerID)
}

func QueueBlindsTimer(interval time.Duration, game *GameInstance) {
	time.Sleep(interval)

	interrupt := GameInterrupt{I_GAME_NEW_BLINDS}

	game.interrupts <- interrupt
}
