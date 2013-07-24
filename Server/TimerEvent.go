package Server

import (
	"time"
)

func QueueTurnTimer(playerID int, timerID int, interval time.Duration, g *GameInstance) {
	time.Sleep(interval)

	action := GameAction{}
	canCheck := g.currentLA.check
	if canCheck {
		action.aType = CHECK
	} else {
		action.aType = FOLD
	}

	g.TakeTurn(playerID, action, true, timerID)
}

func QueueBlindsTimer(interval time.Duration, g *GameInstance) {
	time.Sleep(interval)

	interrupt := GameInterrupt{I_GAME_NEW_BLINDS}

	g.interrupts <- interrupt
}
