package main

import (
	"github.com/jashper/texas-coinem/Server"
	"sync"
)

func main() {
	var db Server.Database
	db.Init("localhost:28015", "test", 10000)

	var entropy Server.EntropyPool
	entropy.Init(10000, 1000)

	var handEval Server.HandEvaluator
	handEval.Init("/Users/jashper/go/bin/resources/texas-coinem/HandRanks.dat")

	context := Server.ServerContext{&db, &entropy, &handEval}

	manager := Server.ConnectionManager{&context}
	go manager.Run("tcp", ":6666")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
