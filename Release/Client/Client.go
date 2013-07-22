package main

import "github.com/jashper/texas-coinem/Client"
import "sync"

func main() {
	connector := Client.Connector{}
	connector.Start("tcp", ":7001")

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
