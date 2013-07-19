package Server

type Interrupt int

const (
	I_GAME_PLAYER_DROP Interrupt = iota
	I_GAME_NEW_BLINDS
)

type GameInterrupt struct {
	iType Interrupt
}
