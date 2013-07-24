package main

import (
	"github.com/jashper/texas-coinem/Server"
	"sync"
	"time"
)

func main() {
	var db Server.Database
	db.Init("localhost:28015", "test", 10000)

	var entropy Server.EntropyPool
	entropy.Init(10000, 1000)

	var handEval Server.HandEvaluator
	handEval.Init("/Users/tanderson/go/bin/resources/texas-coinem/HandRanks.dat")

	var context Server.ServerContext
	context.Init(&db, &entropy, &handEval)

	var manager Server.ConnectionManager
	go manager.Init("tcp", ":7001", &context)

	sbs := []float64{2, 4, 8, 16}
	antes := []float64{0, 1, 2, 4}
	var blinds Server.Blinds
	blinds.Init(sbs, antes)

	var params Server.GameParameters
	levelTime := time.Duration(30) * time.Second
	turnTime := time.Duration(360) * time.Second
	extraTime := time.Duration(0)
	params.Init(Server.HOLDEM, Server.NO_LIMIT, blinds,
		1500, 3, levelTime, turnTime, extraTime)

	var game Server.GameInstance
	context.CurrentGame = &game
	for len(context.Connections) < 3 {
		time.Sleep(1 * time.Second)
	}
	time.Sleep(4 * time.Second)

	game.Init(&context, params, context.Connections)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
