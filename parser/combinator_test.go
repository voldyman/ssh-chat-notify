package parser

import (
	"testing"

	parsec "github.com/prataprc/goparsec"
)

type validateFn func(parsedNode parsec.ParsecNode) bool

func TestLines(t *testing.T) {
	checks := []struct {
		msg      string
		validate validateFn
	}{
		{msg: "chris: it might not merge nicely that way though unless rebasing", validate: validatePubMsg("chris", "it might not merge nicely that way though unless rebasing")},
		{msg: " * gurken joined. (Connected: 12)", validate: validateInfoMessage("gurken", UserJoined)},
		{msg: " * mike left. (After 60 seconds)", validate: validateInfoMessage("mike", UserLeft)},
		{msg: "shazow: what's wrong with ptys?", validate: validatePubMsg("shazow", "what's wrong with ptys?")},
		{msg: "** voldyman has flexible moral values", validate: validateMeMessage("voldyman", "has flexible moral values")},
		{msg: "[PM from Guest91] private message for testing.", validate: validatePrivMessage("Guest91", "private message for testing.")},
		{msg: "[voldyman] some complicated, ardous message", validate: validateEchoMessage("voldyman", "some complicated, ardous message", AckMsgPublic)},
		{msg: "[notifyi] /msg chirs parsing is tough", validate: validateEchoMessage("notifyi", "/msg chirs parsing is tough", AckMsgPublic)},
		{msg: "-> [Sent PM to voldyman]", validate: validateEchoMessage("voldyman", "", AckMsgPrivate)},
		{msg: "-> Message rejected: Rate limiting is in effect.", validate: validateSystemMessage("Message rejected: Rate limiting is in effect.")},
	}
	parser := createLineParser()

	for _, check := range checks {
		result, _ := parser(parsec.NewScanner([]byte(check.msg)))
		if !check.validate(result) {
			t.Fatal("unable to correctly parse:", check.msg)
		}
	}
}

func validatePubMsg(username, message string) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(PublicMsg); ok {
			return result.From == username && result.Message == message
		}
		return false
	}
}

func validateMeMessage(username, message string) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(ActionMsg); ok {
			return result.From == username && result.Message == message
		}
		return false
	}
}

func validateInfoMessage(username string, status UserConnStatus) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(JoinMsg); ok {
			return result.Username == username && result.Status == status
		}
		return false
	}
}

func validatePrivMessage(username, message string) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(PrivateMsg); ok {
			return result.From == username && result.Message == message
		}
		return false
	}
}

func validateEchoMessage(username, message string, typ AckMsgType) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(AckMsg); ok {
			return result.Username == username && result.Message == message && result.Type == typ
		}
		return false
	}
}

func validateSystemMessage(message string) validateFn {
	return func(parsedNode parsec.ParsecNode) bool {
		if result, ok := parsedNode.(SystemMsg); ok {
			return result.Message == message
		}
		return false
	}
}
