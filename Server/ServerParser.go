package Server

import (
	m "github.com/jashper/texas-coinem/Message"
	"reflect"
)

type Parser struct {
	connection *Connection
	context    *ServerContext
}

func (this *Parser) Init(connection *Connection, context *ServerContext) {
	this.connection = connection
	this.context = context
}

func (this *Parser) getData(paramCount int) [][]byte {
	data := make([][]byte, paramCount)

	var length [1]byte
	for i := 0; i < paramCount; i++ {
		this.connection.Read(length[:])
		data[i] = make([]byte, uint8(length[0]))
		this.connection.Read(data[i])
	}

	return data
}

func (this *Parser) sendMessage(params []reflect.Value) {
	message := make([]byte, 0)
	for i := 0; i < len(params); i++ {
		message = append(message, byte(reflect.TypeOf(params[0]).Size()))
		message = append(message, []byte(params[i].)...)
	}
	this.connection.Write(message)
}

func (this *Parser) Message(mType m.ServerMessage) {
	if mType == m.SM_LOGIN_REGISTER {
		data := this.getData(2)
		username := string(data[0])
		password := string(data[1])

		err := RegisterUser(username, password, this.context)
		var message [1]byte
		if err == nil {
			message[0] = byte(m.CM_LOGIN_REGISTER_SUCCESS)
		} else {
			message[0] = byte(m.CM_LOGIN_REGISTER_FAILURE)
		}

		toWrite := append(message[:], uint8(len(username)))
		toWrite = append(toWrite, []byte(username)...)
		toWrite = append(toWrite, uint8(len(password)))
		toWrite = append(toWrite, []byte(password)...)
	}
}
