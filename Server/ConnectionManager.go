package Server

import "net"
import "fmt"

type ConnectionManager struct {
	Context *ServerContext
}

func (this ConnectionManager) Run(network, address string) {
	listener, err := net.Listen(network, address)
	if err != nil {
		fmt.Printf("%s", "CRITICAL : Failed to start ConnectionManager \n")
	}
	fmt.Printf("%s", "Successful start of ConnectionManager \n")

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("%s", "CRITICAL : User failed to connect \n")
			continue
		}
		go this.handleConnection(connection)
	}
}

func (this ConnectionManager) handleConnection(connection net.Conn) {
	defer connection.Close()
	fmt.Printf("%s", "New user connected\n")

	for {
		b := make([]byte, 100) // TODO: Buffer size?
		_, err := connection.Read(b)
		if err != nil {
			fmt.Printf("%s", "CRITICAL : Failed to read from user \n")
			return
		}
		fmt.Printf("%v", string(b)+"\n")
	}
}
