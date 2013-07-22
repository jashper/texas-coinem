package Client

import "net"
import "fmt"

type Connector struct {
	Connection net.Conn
}

func (this *Connector) Start(network, address string) {
	conn, err := net.Dial(network, address)
	this.Connection = conn
	if err != nil {
		fmt.Println("Failed to connect to Server")
		return
	}
	fmt.Println("Successfully connected to Server")

	go this.read()

	for {
		var input string
		fmt.Scanf("%v", &input)
		this.Connection.Write([]byte(input))
	}
}

func (this *Connector) read() {
	for {
		bytes := make([]byte, 100)
		length, err := this.Connection.Read(bytes)
		if err != nil {
			fmt.Println("Server offline - disconnect")
			return
		}

		message := string(bytes[:length])
		fmt.Print(message)
	}
}
