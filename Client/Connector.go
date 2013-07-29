package Client

import "net"
import "fmt"
import m "github.com/jashper/texas-coinem/Message"

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

	/*for {
		var input string
		fmt.Scanf("%v", &input)
		this.socket.Write([]byte(input))
	}*/
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

func (this *Connector) Read(message []byte) { //input slice must be initialized
	this.socket.Read(message)
}
