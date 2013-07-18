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

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
