package Server

import (
	"fmt"
	"net"
)

/*
	This class continually listens to a socket and parses
	any incoming messages.  It also provides a method
	to write messages to a socket.
*/

// ############################################
//     Constructor Struct & Init
// ############################################

type Connection struct {
	socket   net.Conn
	parser   Parser
	userName string
}

func (this *Connection) Init(socket net.Conn, context *ServerContext) {
	this.socket = socket
	this.parser.Init(this, context)
	go this.run()
}

func (this *Connection) run() {
	defer this.socket.Close()
	fmt.Println("New user connected")

	for {
		var message [1]byte
		_, err := this.socket.Read(message[:])
		if err != nil {
			fmt.Println("User disconnected")
			return
		}

		this.parser.Message(ServerMessage(message[0]))
	}
}

func (this *Connection) Write(message []byte) error {
	_, err := this.socket.Write(message)
	return err
}

// only to be used by parser
func (this *Connection) Read(message []byte) { //input slice must be initialized
	this.socket.Read(message)
}
