package Server

import (
	"fmt"
)

type HoldEMGame struct {
	GameInstance
}

func (this *HoldEMGame) UpdateState(playerID int32, command string) {
	fmt.Println("test")
}
