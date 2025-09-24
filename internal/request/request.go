package request

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strings"

	"github.com/nichol20/http-server/internal/header"
)

type parserState string

const (
	Initialized   parserState = "initialized"
	ParsingHeader parserState = "parsing_header"
	Done          parserState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
type Request struct {
	RequestLine RequestLine
	Header      header.Header
	ParserState parserState
}

var allowedMethods = map[string]struct{}{
	"GET":     {},
	"POST":    {},
	"PUT":     {},
	"DELETE":  {},
	"PATCH":   {},
	"OPTIONS": {},
	"HEAD":    {},
}

var supportedVersions = map[string]struct{}{
	//"1.0": {},
	"1.1": {},
}

const INITIAL_BUFFER_SIZE = 1024
const CRLF = "\r\n"

func (r *Request) parse(data []byte) (int, error) {
	if r.ParserState == Done {
		return 0, fmt.Errorf("error: trying to read data in a done state")
	}

	if r.ParserState == Initialized {
		rl, consumed, err := parseRequestLine(data)
		if rl != nil {
			r.RequestLine = *rl
			r.ParserState = ParsingHeader
		}
		return consumed, err
	}

	if r.ParserState == ParsingHeader {
		consumed, done, err := r.Header.Parse(data)
		if done {
			r.ParserState = Done
		}
		return consumed, err
	}

	return 0, fmt.Errorf("error: unkown state")
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	dataStr := string(data)
	if i := strings.Index(dataStr, CRLF); i == -1 {
		return nil, 0, nil
	}

	rlStr := strings.Split(dataStr, CRLF)[0]

	returnErr := func(err error) (*RequestLine, int, error) {
		return nil, 0, err
	}

	parts := strings.Fields(rlStr)
	if len(parts) != 3 {
		return returnErr(ErrMalformedRequestLine)
	}

	method := parts[0]
	if _, ok := allowedMethods[method]; !ok {
		return returnErr(ErrMethodNotAllowed)
	}

	if !strings.HasPrefix(parts[2], "HTTP/") {
		return returnErr(ErrInvalidHTTPVersion)
	}
	httpVersion := strings.TrimPrefix(parts[2], "HTTP/")
	if _, ok := supportedVersions[httpVersion]; !ok {
		return returnErr(ErrUnsupportedHTTPVersion)
	}

	requestTarget := parts[1]

	return &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        method,
	}, len(rlStr) + len(CRLF), nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		RequestLine: RequestLine{},
		Header:      header.NewHeader(),
		ParserState: Initialized,
	}

	buf := make([]byte, INITIAL_BUFFER_SIZE)
	bufIdx := 0

	for request.ParserState != Done {
		n, err := reader.Read(buf[bufIdx:])

		if err != nil && !errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("error reading data: %w", err)
		}

		bufIdx += n
		bufLen := len(buf)
		if bufIdx == bufLen {
			newSize := bufLen * 2
			newBuf := make([]byte, newSize)
			copy(newBuf, buf)
			buf = newBuf
		}

		consumed, err := request.parse(buf[:bufIdx])
		if err != nil {
			return nil, fmt.Errorf("error parsing data: %w", err)
		}

		if consumed > 0 {
			newSize := math.Ceil((float64(bufIdx) / float64(INITIAL_BUFFER_SIZE))) * INITIAL_BUFFER_SIZE
			newBuf := make([]byte, int(newSize))
			copy(newBuf, buf[consumed:])
			buf = newBuf
		}
		bufIdx -= consumed
	}

	return request, nil
}
