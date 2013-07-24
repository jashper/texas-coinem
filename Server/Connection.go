package Server

import (
	"fmt"
	"net"
	"strings"
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
	userName string
	game     *GameInstance
}

func (this *Connection) Init(socket net.Conn, game *GameInstance) {
	this.socket = socket
	this.game = game
	go this.run()
}

func (this *Connection) run() {
	defer this.socket.Close()
	fmt.Println("New user connected")

	for {
		bytes := make([]byte, 100) // TODO: packet size?
		length, err := this.socket.Read(bytes)
		if err != nil {
			fmt.Println("User disconnected")
			return
		}

		message := string(bytes[:length])

		// TODO: parse message
	}
}

func (this *Connection) Write(message string) error {
	_, err := this.socket.Write([]byte(message))
	return err
}
