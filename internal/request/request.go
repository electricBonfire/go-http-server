package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

var ERROR_BAD_REQUEST = fmt.Errorf("bad request")
var SEPARATOR = "\r\n"

func isUpper(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) || !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func (r *RequestLine) ValidHTTP() bool {
	return r.HttpVersion == "HTTP/1.1"
}

func parseRequestLine(line string) (*RequestLine, error) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return nil, errors.Join(ERROR_BAD_REQUEST, fmt.Errorf("Not enough part to the request-line"))
	}

	httpParts := strings.Split(parts[2], "/")
	if len(httpParts) != 2 || httpParts[0] != "HTTP" || httpParts[1] != "1.1" {
		return nil, errors.Join(ERROR_BAD_REQUEST, fmt.Errorf("Invalid HttpVersion"))
	}

	return &RequestLine{
		Method:        parts[0],
		RequestTarget: parts[1],
		HttpVersion:   httpParts[1],
	}, nil

}

func RequestFromReader(reader io.Reader) (*Request, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("Error with io.ReadAll"), err)
	}
	request := string(data)
	index := strings.Index(request, "\r\n")
	if index == -1 {
		return nil, ERROR_BAD_REQUEST
	}

	rl, err := parseRequestLine(request[:index])
	if err != nil {
		return nil, err
	}

	return &Request{
		RequestLine: *rl,
	}, err
}
