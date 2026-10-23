package protocol

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	getRegex = regexp.MustCompile(`^key_in-\s*get\s*\{\s*([^{}:;\s]+)\s*\}$`)
	setRegex = regexp.MustCompile(`^key_in-\s*set\s*\{\s*([^{}:;\s]+)\s*:\s*([^{}:;\s]+)\s*\}$`)
)

type Request struct {
	Command  string
	NrID     string
	Password string
}

func (request *Request) String() string {
	switch request.Command {
	case CmdGet:
		return fmt.Sprintf("key_in-get { %s };", request.NrID)
	case CmdSet:
		return fmt.Sprintf("key_in-set { %s : %s };", request.Password, request.NrID)
	default:
		return ""
	}
}

func ParseRequest(raw string) (*Request, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasSuffix(trimmed, Terminator) {
		return nil, ErrMissingTerminator
	}
	body := strings.TrimSpace(strings.TrimSuffix(trimmed, Terminator))

	if match := getRegex.FindStringSubmatch(body); match != nil {
		return &Request{
			Command: CmdGet,
			NrID:    match[1],
		}, nil
	}

	if match := setRegex.FindStringSubmatch(body); match != nil {
		return &Request{
			Command:  CmdSet,
			Password: match[1],
			NrID:     match[2],
		}, nil
	}

	return nil, ErrMalformedRequest
}
