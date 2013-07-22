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

	manager := Server.ConnectionManager{&context}
	go manager.Run("tcp", ":7001")

	sbs := []float64{2, 4, 8, 16}
	antes := []float64{0, 1, 2, 4}
	var blinds Server.Blinds
	blinds.Init(sbs, antes)

	var params Server.GameParameters
	levelTime := time.Duration(5) * time.Minute
	turnTime := time.Duration(360) * time.Second
	extraTime := time.Duration(0)
	params.Init(Server.HOLDEM, Server.NO_LIMIT, blinds,
		1500, 3, levelTime, turnTime, extraTime)

	for len(context.Connections) < 3 {
		time.Sleep(1 * time.Second)
	}

	time.Sleep(7 * time.Second)

	var game Server.GameInstance
	game.Init(&context, context.Connections, params)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
