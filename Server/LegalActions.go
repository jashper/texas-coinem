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

func (this *GameAction) toString() (value string) {
	if *this == FOLD {
		value = "FOLD"
	} else if *this == CHECK {
		value = "CHECK"
	} else if *this == CALL {
		value = "CALL"
	} else if *this == BET {
		value = "BET"
	} else if *this == RAISE {
		value = "RAISE"
	} else if *this == ALLIN {
		value = "ALLIN"
	}

	return
}
