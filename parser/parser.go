package parser

import (
	"fmt"

	parsec "github.com/prataprc/goparsec"
)

// Parser is used to understand lines sent by ssh-chat
type Parser struct {
	lineParser parsec.Parser
}

// New creates a new parser
func New() *Parser {
	return &Parser{
		lineParser: createLineParser(),
	}
}

// Parse published line
func (p *Parser) Parse(line []byte) (RoomMsg, error) {
	presult, _ := p.lineParser(parsec.NewScanner(line))
	if presult == nil {
		return nil, fmt.Errorf("unable to parse line '%s'", line)
	}
	return presult, nil
}
