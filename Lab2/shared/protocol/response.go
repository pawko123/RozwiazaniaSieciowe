package protocol

import (
	"fmt"
	"regexp"
	"strings"
)

var respRegex = regexp.MustCompile(`^key_out-\s*\{\s*([^{}:;\s]*)\s*:\s*(yes|not|error)\s*\}\s*;?$`)

type Response struct {
	NrID     string
	Contents string
}

func (r Response) String() string {
	if r.NrID == "" {
		return fmt.Sprintf("key_out-{:%s};", r.Contents)
	}
	return fmt.Sprintf("key_out-{%s : %s};", r.NrID, r.Contents)
}

func ParseResponse(raw string) (*Response, error) {
	trimmed := strings.TrimSpace(raw)
	match := respRegex.FindStringSubmatch(trimmed)
	if match == nil {
		return nil, ErrMalformedResponse
	}
	return &Response{
		NrID:     match[1],
		Contents: match[2],
	}, nil
}
