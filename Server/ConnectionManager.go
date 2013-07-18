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
	}
	fmt.Println("Successful start of ConnectionManager")

	for {
		socket, err := listener.Accept()
		if err != nil {
			fmt.Println("CRITICAL : User failed to connect")
			continue
		}
		connection := Connection{socket}
		go connection.Handle()
	}
}
