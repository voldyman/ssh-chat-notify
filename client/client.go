package client

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/lunixbochs/vtclean"
	"github.com/pkg/errors"
	"github.com/shazow/rateio"
	"golang.org/x/crypto/ssh"
)

const noColor = false
const discardEcho = true

// Client is used to communicate with ssh-chat or other ssh-sessions
type Client struct {
	conn    net.Conn
	client  *ssh.Client
	session *ssh.Session

	scanner *bufio.Scanner
	writer  io.Writer

	err error

	ratelimit rateio.Limiter
}

// CreateClient establishes a connections with the destination as the given username
func CreateClient(destination, username string) (*Client, error) {
	client, conn, err := createSSHClient(destination, username)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, errors.Wrap(err, "unable to create ssh client")
	}

	session, err := client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "unable to establish ssh session")
	}

	r, w, err := createSessionIO(session)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open read/write connection to the session")
	}

	return &Client{
		conn:      conn,
		client:    client,
		session:   session,
		scanner:   bufio.NewScanner(r),
		writer:    w,
		ratelimit: rateio.NewSimpleLimiter(3, time.Second*3),
	}, nil
}

func createSSHClient(dest, username string) (*ssh.Client, net.Conn, error) {
	signer, err := getSigner()
	if err != nil {
		signer, err = genSinger()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "unable to get signer and then generate it")
		}
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	conn, err := net.DialTimeout("tcp", dest, config.Timeout)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to estable tcp connection to ssh server")
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, dest, config)
	if err != nil {
		return nil, conn, errors.Wrap(err, "unable to create ssh client conn")
	}
	client := ssh.NewClient(c, chans, reqs)

	return client, conn, nil
}

// ScanLine reads the connection till the next new line
func (c *Client) ScanLine() (string, error) {
	if c.err != nil {
		return "", c.err
	}
	// c.conn.SetReadDeadline(time.Now().Add(30 * time.Minute))

	continueReading := c.scanner.Scan()
	if c.scanner.Err() != nil {
		c.err = c.scanner.Err()
		return "", c.err
	}
	if !continueReading {
		c.err = io.EOF
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
	err := c.session.Close()
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "unable to close underlying ssh session")
	}
	err = c.client.Close()
	if err != nil {
		return errors.Wrap(err, "unable to close underlying ssh client")
	}
	return nil
}

func createSessionIO(session *ssh.Session) (io.Reader, io.WriteCloser, error) {
	w, err := session.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	r, err := session.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}

	err = session.Shell()
	if err != nil {
		return nil, nil, err
	}

	err = session.RequestPty("xterm", 80, 40, ssh.TerminalModes{})
	return r, w, err
}

func getSigner() (ssh.Signer, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to find user's home dir for locating ssh key: %w", err)
	}

	key, err := ioutil.ReadFile(homeDir + "/.ssh/id_rsa")
	if err != nil {
		return nil, fmt.Errorf("unable to read local ssh key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	return signer, nil
}

func genSinger() (ssh.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2014)
	if err != nil {
		return nil, fmt.Errorf("unable to generate rsa key: %w", err)
	}
	return ssh.NewSignerFromKey(key)

}
