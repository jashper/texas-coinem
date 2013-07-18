package Server

type GameVariant int

const (
	HOLDEM GameVariant = iota
)

type GameLimit int

const (
	FIXED_LIMIT GameLimit = iota
	POT_LIMIT
	NO_LIMIT
)

type GameParameters struct {
	Variant     GameVariant
	Limit       GameLimit
	Blinds      Blinds
	ChipCount   int
	PlayerCount int
	LevelTime   int
	TurnTime    int
	ExtraTime   int
}

func (this *GameParameters) Init(variant GameVariant, limit GameLimit, blinds Blinds,
	chipCount, playerCount, levelTime, turnTime, extraTime int) {

	this.Variant = variant
	this.Limit = limit
	this.Blinds = blinds
	this.ChipCount = chipCount
	this.PlayerCount = playerCount
	this.LevelTime = levelTime
	this.TurnTime = turnTime
	this.ExtraTime = extraTime
}
