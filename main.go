package main

import (
	"fmt"
	"os"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"github.com/voldyman/ssh-chat-notify/client"
	"github.com/voldyman/ssh-chat-notify/notifyi"
	"github.com/voldyman/ssh-chat-notify/parser"
)

const username = "notifyi"

const sshChatHost = "localhost:2022"

var logger log.Logger

func main() {
	logger = golog.New(os.Stderr, log.Debug)

	if err := run(); err != nil {
		logger.Error("Failed:", err)
	}
}

type clientComms struct {
	client *client.Client
}

func (c *clientComms) PrivateMessage(toUsername, message string) error {
	c.client.WriteLine(fmt.Sprintf("/msg %s %s", toUsername, message))
	return nil
}

func (c *clientComms) PublicMessage(message string) error {
	c.client.WriteLine(message)
	return nil
}

func run() error {
	dest := sshChatHost
	if len(os.Args) >= 2 {
		dest = os.Args[1]
	}

	client, err := client.CreateClient(dest, username)
	if err != nil {
		return err
	}
	defer client.Close()

	bot := notifyi.New(username, &clientComms{client})
	lineParser := parser.New()

	for {
		line, err := client.ScanLine()
		if err != nil {
			return err
		}

		fmt.Println("Got Line", line)
		parsedResult, err := lineParser.Parse([]byte(line))
		if err != nil {
			logger.Warningf("parsing failed: %+v", err)
			continue
		}

		switch result := parsedResult.(type) {
		case parser.PrivateMsg:
			logger.Info("Private Message", quote(result.From), "->", result.Message)
			bot.PrivateMessage(result.From, result.Message)

		case parser.PublicMsg:
			logger.Info("Public Message", quote(result.From), "->", result.Message)
			bot.PublicMessage(result.From, result.Message)

		case parser.ActionMsg:
			logger.Info("Smartass says:", quote(result.From), "->", result.Message)
			bot.ActionMessage(result.From, result.Message)

		case parser.UsernameChangeMsg:
			logger.Info("Nick change", quote(result.FromUsername), "->", result.ToUsername)
			bot.UsernameChangeMessage(result.FromUsername, result.ToUsername)

		case parser.JoinMsg:
			switch result.Status {
			case parser.UserJoined:
				bot.UserJoinedMessage(result.Username)
			case parser.UserLeft:
				bot.UserLeftMessage(result.Username)
			}
			logger.Info("User", quote(result.Username), "has", result.Status)

		}

	}
}

func quote(s string) string {
	return "\"" + s + "\""
}
