package client

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

func createSession(dest, username string) (*ssh.Session, error) {
	signer, err := getSigner()
	if err != nil {
		signer, err = genSinger()
		if err != nil {
			return nil, err
		}
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	conn, err := ssh.Dial("tcp", dest, config)
	if err != nil {
		return nil, err
	}

	session, err := conn.NewSession()
	if err != nil {
		return nil, err
	}

	return session, nil
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
