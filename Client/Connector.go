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
		fmt.Printf("%s", "CRITICAL : Failed to connect to Server \n")
		return
	}
	fmt.Printf("%s", "Successfully connected to Server \n")
}
