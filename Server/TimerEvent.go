package Server

import (
	"time"
)

func QueueHandTimer(playerID int32, timerID int, interval time.Duration, game *GameInstance) {
	time.Sleep(interval * time.Second)

	game.TakeTurn(playerID, "", true, timerID)
}
