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
}
