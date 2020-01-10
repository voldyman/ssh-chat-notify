package main

import (
	"os"

	"github.com/alexcesaro/log"
	"github.com/alexcesaro/log/golog"
	"github.com/voldyman/ssh-chat-notify/client"
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

	lineParser := parser.New()

	for {
		line, err := client.ScanLine()
		if err != nil {
			return err
		}

		parsedResult, err := lineParser.Parse([]byte(line))
		if err != nil {
			logger.Warningf("parsing failed: %+v", err)
			continue
		}

		switch result := parsedResult.(type) {
		case parser.UsernameChangeMsg:
			logger.Info("Nick change", quote(result.FromUsername), "->", result.ToUsername)

		case parser.JoinMsg:
			logger.Info("User", quote(result.Username), "has", result.Status)

		case parser.PrivateMsg:
			logger.Info("Private Message", quote(result.From), "->", result.Message)
			client.WriteLine("stop tickeling me " + result.From)

		case parser.PublicMsg:
			logger.Info("Public Message", quote(result.From), "->", result.Message)

		case parser.ActionMsg:
			logger.Info("Smartass says:", quote(result.From), "->", result.Message)
		}

	}
}

func quote(s string) string {
	return "\"" + s + "\""
}
