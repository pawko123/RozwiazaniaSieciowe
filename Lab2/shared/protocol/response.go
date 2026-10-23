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

func (response *Response) String() string {
	if response.NrID == "" {
		return fmt.Sprintf("key_out-{:%s};", response.Contents)
	}
	return fmt.Sprintf("key_out-{%s : %s};", response.NrID, response.Contents)
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
