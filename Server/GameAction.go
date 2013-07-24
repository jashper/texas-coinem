package Server

/*
	This class pairs a constant enum representing a legal game action
	with an optional value.
*/

// ############################################
//     Enum definition
// ############################################

type GameActionType int

const (
	FOLD GameActionType = iota
	CHECK
	CALL
	BET
	RAISE
	ALLIN
)

// Helper parsing method
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

// ############################################
//     Constructor Struct
// ############################################

type GameAction struct {
	aType GameActionType
	value float64
}
