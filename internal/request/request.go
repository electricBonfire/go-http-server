package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unicode"
)

type parserState string

type Request struct {
	RequestLine RequestLine
	state       parserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const (
	StateInit parserState = "init"
	StateDone parserState = "done"
)

var ERROR_BAD_REQUEST = fmt.Errorf("bad request")
var SEPARATOR = []byte("\r\n")

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

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	index := bytes.Index(b, SEPARATOR)
	if index == -1 {
		return nil, 0, nil
	}

	line := b[:index]
	read := index + len(SEPARATOR)

	parts := bytes.Split(line, []byte(" "))
	if len(parts) != 3 {
		return nil, index, errors.Join(ERROR_BAD_REQUEST, fmt.Errorf("Not enough part to the request-line"))
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" || string(httpParts[1]) != "1.1" {
		return nil, index, errors.Join(ERROR_BAD_REQUEST, fmt.Errorf("Invalid HttpVersion"))
	}

	return &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}, read, nil

}

func (r *Request) parse(data []byte) (int, error) {
	read := 0
outer:
	for {
		switch r.state {
		case StateInit:
			rl, n, err := parseRequestLine(data[read:])
			if err != nil {
				return 0, err
			}
			if n == 0 {
				break outer
			}

			r.RequestLine = *rl
			read += n

			r.state = StateDone

		case StateDone:
			break outer
		}
	}
	return read, nil
}

func (r *Request) isDone() bool {
	return r.state == StateDone
}

func newRequest() *Request {
	return &Request{
		state: StateInit,
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()

	buf := make([]byte, 1024)
	bufLen := 0

	for !request.isDone() {
		n, err := reader.Read(buf[bufLen:])
		if err != nil {
			// TODO: How do we want to handle errors?
			return nil, err
		}
		bufLen += n
		readN, err := request.parse(buf[:bufLen])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[readN:bufLen])
	}

	return request, nil
}
