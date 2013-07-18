package main

import (
	//"fmt"
	"github.com/jashper/texas-coinem/Server"
	"sync"
)

func main() {
	context := Server.ServerContext{}

	manager := Server.ConnectionManager{&context}
	go manager.Run("tcp", ":6666")

	game := Server.HoldEMGame{}
	game.TakeTurn(0, "", false, 0)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
