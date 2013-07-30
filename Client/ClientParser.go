package Client

import (
	"fmt"
	m "github.com/jashper/texas-coinem/Message"
)

type Parser struct {
	connection *Connector
}

func (this *Parser) Init(connection *Connector) {
	this.connection = connection
}

func (this *Parser) GetStrings(paramCount int) []string {
	data := make([]string, paramCount)

	var length [1]byte
	for i := 0; i < paramCount; i++ {
		this.connection.Read(length[:])
		temp := make([]byte, int(length[0]))
		this.connection.Read(temp)
		data[i] = string(temp)
	}

	return data
}

func (this *Parser) AppendStrings(message *[]byte, params []string) {
	for i := 0; i < len(params); i++ {
		str := params[i]
		if len(str) > 255 {
			fmt.Println("CRITICAL : Tried to send a string message greater than 255 characters")
			return
		}
		*message = append(*message, byte(len(str)))
		*message = append(*message, []byte(str)...)
	}
}

func (this *Parser) Message(mType m.ClientMessage) {
	if mType == m.CM_LOGIN_REGISTER_SUCCESS {
		fmt.Println("Registration Success")
	} else if mType == m.CM_LOGIN_REGISTER_DUPLICATE {
		fmt.Println("Registration Duplicate")
	}
}
