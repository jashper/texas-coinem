package Client

import (
	"fmt"
	m "github.com/jashper/texas-coinem/Message"
	r "reflect"
)

type Parser struct {
	connection *Connector
}

func (this *Parser) Init(connection *Connector) {
	this.connection = connection
}

func (this *Parser) getData(paramCount int) [][]byte {
	data := make([][]byte, paramCount)

	var length [1]byte
	for i := 0; i < paramCount; i++ {
		this.connection.Read(length[:])
		data[i] = make([]byte, int(length[0]))
		this.connection.Read(data[i])
	}

	return data
}

func (this *Parser) sendMessage(mType m.ServerMessage, params []r.Value) {
	message := make([]byte, 1)
	message[0] = byte(mType)
	for i := 0; i < len(params); i++ {
		p := params[0]
		if p.Kind() == r.String {
			pStr := p.String()
			message = append(message, byte(len(pStr)))
			if len(pStr) > 255 {
				fmt.Println("CRITICAL : Tried to send a string message greater than 255 characters")
				return
			}
			message = append(message, []byte(pStr)...)
		} else {
			// TODO: other types
		}
	}
	this.connection.Write(message)
}

func (this *Parser) Message(mType m.ClientMessage) {
	if mType == m.CM_LOGIN_REGISTER_SUCCESS {

	}
}
