package Server

type GameActionType int

const (
	FOLD GameActionType = iota
	CHECK
	CALL
	BET
	RAISE
	ALLIN
)

func (this *GameActionType) toString() (value string) {
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

type GameAction struct {
	aType GameActionType
	value float64
}

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
