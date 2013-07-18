package Server

import "net"
import "fmt"

type Connection struct {
	socket net.Conn
}

func (this *Connection) Handle() {
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
