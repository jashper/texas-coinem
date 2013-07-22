package Server

import "net"
import "fmt"
import "strings"

type Connection struct {
	socket   net.Conn
	userName string
}

func (this *Connection) Init(socket net.Conn) {
	this.socket = socket
	go this.run()
}

func (this *Connection) run() {
	defer this.socket.Close()
	fmt.Println("New user connected")

	for {
		bytes := make([]byte, 100) // TODO: Buffer size?
		length, err := this.socket.Read(bytes)
		if err != nil {
			fmt.Println("User disconnected")
			return
		}

		message := string(bytes[:length])
		if strings.Contains(message, "username") {
			split := strings.Split(message, ":")
			this.userName = split[1]
		}
	}
}

func (this *Connection) Write(message string) error {
	_, err := this.socket.Write([]byte(message))
	return err
}
