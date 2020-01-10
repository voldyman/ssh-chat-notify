package parser

import (
	"strings"

	parsec "github.com/prataprc/goparsec"
)

const usernameRegex = `[\w.-]*`

func createLineParser() parsec.Parser {
	infoParser := createInfoParser()
	messageParser := createMessageParser()
	meMessageParser := createMeParser()

	return parsec.OrdChoice(selectFirstNode, infoParser, meMessageParser, messageParser)
}

func createInfoParser() parsec.Parser {
	userNameTok := parsec.Token(usernameRegex, "USERNAME")
	joinStatusParser := parsec.And(statusIfMatch(UserJoined), parsec.Atom("joined.", "JOIN_STATUS"), parsec.Token(".*", "IGNORE"))

	leftStatusParser := parsec.And(statusIfMatch(UserLeft), parsec.Atom("left.", "LEFT_STATUS"), parsec.Token(".*", "IGNORE"))

	membershipStatusParser := parsec.OrdChoice(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		return JoinMsg{
			Status: nodes[0].(UserConnStatus),
		}
	}, joinStatusParser, leftStatusParser)

	newNickTok := parsec.Token(usernameRegex, "NEW_USERNAME")
	nickChangeParser := parsec.And(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		nick := nodes[1].(*parsec.Terminal).GetValue()
		// the joined method has a '.' at the end
		nick = strings.TrimSuffix(nick, ".")
		return UsernameChangeMsg{
			ToUsername: nick,
		}
	}, parsec.Atom("is now known as", "_KNOWN_AS"), newNickTok)

	infoSuffixParser := parsec.OrdChoice(selectFirstNode, nickChangeParser, membershipStatusParser)

	infoParser := parsec.And(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		// info_prefixNode := nodes[0]
		usernameNode := nodes[1]
		actionNode := nodes[2]

		username := usernameNode.(*parsec.Terminal).GetValue()
		switch actionMsg := actionNode.(type) {
		case JoinMsg:
			actionMsg.Username = username
			return actionMsg
		case UsernameChangeMsg:
			actionMsg.FromUsername = username
			return actionMsg
		}
		return nil
	}, parsec.Atom("*", "INFO_PREFIX"), userNameTok, infoSuffixParser)

	return infoParser
}

func statusIfMatch(status UserConnStatus) parsec.Nodify {
	return func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		return status
	}
}

func selectFirstNode(nodes []parsec.ParsecNode) parsec.ParsecNode {
	return nodes[0]
}

func createMessageSuffixParser() parsec.Parser {
	msgTok := parsec.Token(`.*`, "MSG")
	return parsec.ManyUntil(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		messageParts := []string{}
		for _, node := range nodes {
			val := node.(*parsec.Terminal).GetValue()
			messageParts = append(messageParts, val)
		}
		// parsec tokenizes by space, need to put them back
		return strings.Join(messageParts, " ")
	}, msgTok, parsec.End())
}

func createMessageParser() parsec.Parser {
	userNameTok := parsec.Token(usernameRegex, "USERNAME")
	msgPartsParser := createMessageSuffixParser()

	pmParser := parsec.And(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		username := nodes[1].(*parsec.Terminal).GetValue()
		message := nodes[3].(string)
		return PrivateMsg{
			From:    username,
			Message: message,
		}
	}, parsec.Atom("[PM from ", "PM_PREFIX"), userNameTok, parsec.Atom("] ", "PM_SUFFIX"), msgPartsParser)
	msgParser := parsec.And(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		username := nodes[0].(*parsec.Terminal).GetValue()
		message := nodes[2].(string)
		return PublicMsg{
			From:    username,
			Message: message,
		}
	}, userNameTok, parsec.Atom(":", "MSG_SEP"), msgPartsParser)

	return parsec.OrdChoice(selectFirstNode, pmParser, msgParser)
}

func createMeParser() parsec.Parser {
	usernameParser := parsec.Token(usernameRegex, "USERNAME")
	return parsec.And(func(nodes []parsec.ParsecNode) parsec.ParsecNode {
		username := nodes[1].(*parsec.Terminal).GetValue()
		message := nodes[2].(string)
		return ActionMsg{
			From:    username,
			Message: message,
		}
	}, parsec.Atom("**", "PREFIX"), usernameParser, createMessageSuffixParser())
}
