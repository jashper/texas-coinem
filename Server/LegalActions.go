package Server

type GameAction int

const (
	FOLD GameAction = iota
	CHECK
	CALL
	BET
	RAISE
	ALLIN
)

type LegalActions struct {
	fold  bool
	check bool
	call  bool
	bet   bool
	raise bool
	allin bool

	min float64
	max float64
}
