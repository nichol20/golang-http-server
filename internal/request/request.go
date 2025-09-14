package request

import (
	"fmt"
	"io"
	"strings"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("error reading request: %w", err)
	}
	requestStr := string(buf)

	splitedReq := strings.Split(requestStr, "\r\n")

	if len(splitedReq) == 0 {
		return nil, fmt.Errorf("request is empty")
	}

	requestLine, err := parseRequestLine(splitedReq[0])
	if err != nil {
		return nil, err
	}

	request := &Request{
		RequestLine: *requestLine,
	}

	return request, nil
}

func parseRequestLine(requestLineStr string) (*RequestLine, error) {
	requestLineParts := strings.Split(requestLineStr, " ")

	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}

	method := requestLineParts[0]
	requestTarget := requestLineParts[1]

	httpVersionParts := strings.Split(requestLineParts[2], "/")
	if len(httpVersionParts) != 2 {
		return nil, fmt.Errorf("invalid http version")
	}

	return &RequestLine{
		HttpVersion:   httpVersionParts[1],
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}
