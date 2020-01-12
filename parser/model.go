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

// AckMsg represents the message sent back by ssh-chat
type AckMsg struct {
	Username string
	Message  string
	Type     AckMsgType
}

// AckMsgType represents what kind of message was acknowledged
type AckMsgType int

const (
	// AckMsgPublic means public message was acknowledged
	AckMsgPublic AckMsgType = iota
	// AckMsgPrivate means private message was acknowledged
	AckMsgPrivate
)

// SystemMsg represents the system messages ssh-chat sends periodically
type SystemMsg struct {
	Message string
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
