package Server

import (
	"fmt"
	"net"
)

/*
	This class listens for new clients and spawns a Connection
	for each one that successfully connects
*/

// ############################################
//     Constructor Struct & Init
// ############################################

type ConnectionManager struct {
	context *ServerContext
}

func (this *ConnectionManager) Init(network, address string, context *ServerContext) {
	this.context = context

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
		c.Init(socket, this.context.CurrentGame)
		this.context.Connections = append(this.context.Connections, &c)
	}
}
