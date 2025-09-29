package request

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/nichol20/http-server/internal/header"
)

type parserState string

const (
	Initialized   parserState = "initialized"
	ParsingHeader parserState = "parsing_header"
	ParsingBody   parserState = "parsing_body"
	Done          parserState = "done"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}
type Request struct {
	ParserState parserState
	RequestLine RequestLine
	Header      header.Header
	Body        []byte
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
	switch r.ParserState {
	case Initialized:
		rl, consumed, err := parseRequestLine(data)
		if rl != nil {
			r.RequestLine = *rl
			r.ParserState = ParsingHeader
		}
		return consumed, err

	case ParsingHeader:
		consumed, done, err := r.Header.Parse(data)
		if done {
			// I'm assuming that if there is no content-length there will be no body
			if len(r.Header.Get("content-length")) == 0 {
				r.ParserState = Done
			} else {
				r.ParserState = ParsingBody
			}
		}
		return consumed, err

	case ParsingBody:
		contentLen, err := strconv.Atoi(r.Header.Get("content-length"))
		if err != nil {
			return 0, fmt.Errorf("invalid content length")
		}
		r.Body = append(r.Body, data...)
		if len(r.Body) > contentLen {
			return 0, fmt.Errorf("the body size is larger than the content length specified in the header")
		}
		if len(r.Body) == contentLen {
			r.ParserState = Done
		}
		return len(data), nil

	case Done:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unkown state")
	}
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
		ParserState: Initialized,
		RequestLine: RequestLine{},
		Header:      header.NewHeader(),
		Body:        []byte{},
	}

	buf := make([]byte, INITIAL_BUFFER_SIZE)
	bufIdx := 0
	reachedEOF := false
	for request.ParserState != Done {
		if reachedEOF {
			return nil, fmt.Errorf("the request parse has not finished, but there is no more data to read")
		}
		n, err := reader.Read(buf[bufIdx:])
		reachedEOF = errors.Is(err, io.EOF)
		if err != nil {
			if !reachedEOF {
				return nil, fmt.Errorf("error reading data: %w", err)
			}
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
