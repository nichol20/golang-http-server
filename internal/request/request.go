package request

import (
	"fmt"
	"io"
	"net/http"
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

var allowedMethods = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodDelete:  {},
	http.MethodPatch:   {},
	http.MethodOptions: {},
	http.MethodHead:    {},
}

var supportedVersions = map[string]struct{}{
	//"1.0": {},
	"1.1": {},
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
	parts := strings.Fields(requestLineStr)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line")
	}

	method := parts[0]
	if _, ok := allowedMethods[method]; !ok {
		return nil, fmt.Errorf("method not allowed")
	}

	requestTarget := parts[1]
	if requestTarget == "" {
		return nil, fmt.Errorf("empty request-target")
	}

	// HTTP-version must start with "HTTP/"
	if !strings.HasPrefix(parts[2], "HTTP/") {
		return nil, fmt.Errorf("invalid http version")
	}
	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if _, ok := supportedVersions[httpVersion]; !ok {
		return nil, fmt.Errorf("http version %q not supported", httpVersion)
	}

	return &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        method,
	}, nil
}
