package client

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/lunixbochs/vtclean"
	"github.com/shazow/rateio"
	"golang.org/x/crypto/ssh"
)

const noColor = false
const discardEcho = true

// Client is used to communicate with ssh-chat or other ssh-sessions
type Client struct {
	sshSession *ssh.Session

	scanner *bufio.Scanner
	writer  io.Writer

	err error

	ratelimit rateio.Limiter
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
		ratelimit:  rateio.NewSimpleLimiter(3, time.Second*3),
	}, nil
}

// ScanLine reads the connection till the next new line
func (c *Client) ScanLine() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	continueReading := c.scanner.Scan()
	if !continueReading {
		c.err = io.EOF
		return "", c.err
	}

	if c.scanner.Err() != nil {
		c.err = c.scanner.Err()
		return "", c.err
	}
	cleanedLine := vtclean.Clean(c.scanner.Text(), noColor)
	return cleanedLine, nil
}

// WriteLine send the given line to ssh-chat
func (c *Client) WriteLine(line string) error {
	if c.ratelimit.Count(1) != nil {
		time.Sleep(1 * time.Second)
	}
	return c.writeLine(line)
}

func (c *Client) writeLine(line string) error {
	_, err := c.writer.Write([]byte(line + "\r\n"))
	if err != nil {
		return fmt.Errorf("unable to write line: %w", err)
	}
	return nil
}

// Close disconnects the client
func (c *Client) Close() error {
	err := c.sshSession.Close()
	if err != nil {
		return fmt.Errorf("unable to close underlying ssh session: %w", err)
	}
	return nil
}
