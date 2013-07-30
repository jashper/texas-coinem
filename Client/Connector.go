package Client

import (
	"fmt"
	m "github.com/jashper/texas-coinem/Message"
	"net"
	"strings"
)

type Connector struct {
	socket net.Conn
	parser Parser
}

func (this *Connector) Start(network, address string) {
	conn, err := net.Dial(network, address)
	this.socket = conn
	this.parser.Init(this)
	if err != nil {
		fmt.Println("Failed to connect to Server")
		return
	}
	fmt.Println("Successfully connected to Server")

	go this.read()

	for {
		var input string
		fmt.Scanf("%v", &input)

		if strings.Contains(input, "register") {
			split := strings.Split(input, ";")
			message := []byte{byte(m.SM_LOGIN_REGISTER)}
			this.parser.AppendStrings(&message, split[1:])
			this.Write(message)
		}
	}
}

func (this *Connector) read() {
	for {
		var message [1]byte
		_, err := this.socket.Read(message[:])
		if err != nil {
			fmt.Println("Server offline - disconnect")
			return
		}

		this.parser.Message(m.ClientMessage(message[0]))
	}
}

func (this *Connector) Write(message []byte) error {
	_, err := this.socket.Write(message)
	return err
}

// only to be used by parser
func (this *Connector) Read(message []byte) { //input slice must be initialized
	this.socket.Read(message)
}
