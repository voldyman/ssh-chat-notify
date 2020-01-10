package parser

// RoomMsg is one of {UsernameChangeMsg, JoinMsg, PrivateMsg, PublicMsg, ActionMsg}
type RoomMsg interface{}

// PublicMsg represents message sent to the room
type PublicMsg struct {
	From    string
	Message string
}

// PrivateMsg represents message sent to the user directly
type PrivateMsg struct {
	From    string
	Message string
}

// ActionMsg represents a public message sent as an action '/me <something>'
type ActionMsg struct {
	From    string
	Message string
}

// JoinMsg represents message published by server about people joining or leaving
type JoinMsg struct {
	Username string
	Status   UserConnStatus
}

// UsernameChangeMsg represents message published by server about users changing their names
type UsernameChangeMsg struct {
	FromUsername string
	ToUsername   string
}

// UserConnStatus represents user's connected status
type UserConnStatus int

const (
	// UserJoined means the user connected to the server
	UserJoined UserConnStatus = iota
	// UserLeft means the user disconnected from the server
	UserLeft
)

func (u UserConnStatus) String() string {
	switch u {
	case UserJoined:
		return "joined"
	case UserLeft:
		return "left"
	}
	return "undef"
}
