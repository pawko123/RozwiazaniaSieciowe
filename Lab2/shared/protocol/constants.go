package protocol

import "errors"

const (
	Terminator   = ";"
	CmdGet       = "get"
	CmdSet       = "set"
	ContentYes   = "yes"
	ContentNot   = "not"
	ContentError = "error"
	DefaultPort  = 2011
)

var (
	ErrMalformedRequest  = errors.New("malformed key protocol request")
	ErrMalformedResponse = errors.New("malformed key protocol response")
	ErrMissingTerminator = errors.New("missing terminating semicolon")
)
