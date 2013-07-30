package Message

type ServerMessage byte

const (
	// Login messages
	SM_LOGIN_REGISTER ServerMessage = iota //username, password string
)

type ClientMessage byte

const (
	// Login messages
	CM_LOGIN_REGISTER_SUCCESS ClientMessage = iota
	CM_LOGIN_REGISTER_DUPLICATE
)
