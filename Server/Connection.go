package Server

import "net"
import "fmt"

type Connection struct {
	socket net.Conn
}

func (this *Connection) Init(socket net.Conn) {
	this.socket = socket
	go this.run()
}

func (this *Connection) run() {
	defer this.socket.Close()
	fmt.Println("New user connected")

	for {
		b := make([]byte, 100) // TODO: Buffer size?
		_, err := this.socket.Read(b)
		if err != nil {
			fmt.Println("User disconnected")
			return
		}
	}
}

//TODO: add write function
