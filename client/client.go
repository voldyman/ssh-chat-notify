package client

import (
	"bufio"
	"fmt"
	"io"

	"github.com/lunixbochs/vtclean"
	"golang.org/x/crypto/ssh"
)

const noColor = false
const discardEcho = true

// Client is used to communicate with ssh-chat or other ssh-sessions
type Client struct {
	sshSession *ssh.Session

	scanner *bufio.Scanner
	writer  io.Writer
}

// CreateClient establishes a connections with the destination as the given username
func CreateClient(destination, username string) (*Client, error) {
	sshSession, err := createSession(destination, username)
	if err != nil {
		return nil, fmt.Errorf("unable to create session: %w", err)
	}
	r, w, err := createSessionIO(sshSession)
	if err != nil {
		return nil, fmt.Errorf("unable to open read/write connection to the session: %w", err)
	}
	return &Client{
		sshSession: sshSession,
		scanner:    bufio.NewScanner(r),
		writer:     w,
	}, nil
}

// ScanLine reads the connection till the next new line
func (s *Client) ScanLine() (string, error) {
	continueReading := s.scanner.Scan()
	if !continueReading {
		return "", s.scanner.Err()
	}
	cleanedLine := vtclean.Clean(s.scanner.Text(), noColor)
	return cleanedLine, nil
}

// WriteLine send the given line to ssh-chat
func (s *Client) WriteLine(line string) error {
	_, err := s.writer.Write([]byte(line + "\r\n"))
	if err != nil {
		return fmt.Errorf("unable to write line: %w", err)
	}
	// discard the echo
	if discardEcho {
		s.ScanLine()
	}
	return nil
}

// Close disconnects the client
func (s *Client) Close() error {
	err := s.sshSession.Close()
	if err != nil {
		return fmt.Errorf("unable to close underlying ssh session: %w", err)
	}
	return nil
}
