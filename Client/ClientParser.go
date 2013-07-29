package Client

import (
	m "github.com/jashper/texas-coinem/Message"
	"net"
)

type Parser struct {
	socket *net.Conn
}

func (this *Parser) Init(socket *net.Conn) {
	this.socket = socket
}

func (this *Parser) getData(paramCount int) [][]byte {
	data := make([][]byte, paramCount)

	var length [1]byte
	for i := 0; i < paramCount; i++ {
		(*this.socket).Read(length[:])
		data[i] = make([]byte, uint8(length[0]))
		(*this.socket).Read(data[i])
	}

	return data
}

func (this *Parser) Message(mType m.ClientMessage) {

}
