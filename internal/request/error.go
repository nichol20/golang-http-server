package request

import "errors"

var (
	ErrMalformedRequestLine   = errors.New("malformed request line")
	ErrInvalidHTTPVersion     = errors.New("invalid http version")
	ErrUnsupportedHTTPVersion = errors.New("unsupported http version")
	ErrMethodNotAllowed       = errors.New("method not allowed")
)
