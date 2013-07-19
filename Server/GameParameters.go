package Server

import (
	"time"
)

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
	ChipCount   float64
	PlayerCount int
	LevelTime   int
	TurnTime    time.Duration
	ExtraTime   time.Duration
}

func (this *GameParameters) Init(variant GameVariant, limit GameLimit, blinds Blinds,
	chipCount float64, playerCount, levelTime int, turnTime, extraTime time.Duration) {

	this.Variant = variant
	this.Limit = limit
	this.Blinds = blinds
	this.ChipCount = chipCount
	this.PlayerCount = playerCount
	this.LevelTime = levelTime
	this.TurnTime = turnTime
	this.ExtraTime = extraTime
}
