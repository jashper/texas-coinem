package Server

import "net"
import "fmt"

type ConnectionManager struct {
	Context *ServerContext
}

func (this ConnectionManager) Run(network, address string) {
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Println("CRITICAL : Failed to start ConnectionManager")
		return
	}
	fmt.Println("Successful start of ConnectionManager")

	for {
		socket, err := listener.Accept()
		if err != nil {
			fmt.Println("CRITICAL : User failed to connect")
			continue
		}
		var c Connection
		c.Init(socket)
		this.Context.Connections = append(this.Context.Connections, &c)
	}
}
