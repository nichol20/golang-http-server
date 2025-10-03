package response

import (
	"fmt"
	"io"

	"github.com/nichol20/http-server/internal/header"
)

type StatusCode uint16

const (
	OK                    StatusCode = 200
	BAD_REQUEST           StatusCode = 400
	INTERNAL_SERVER_ERROR StatusCode = 500
)

var reasonPhrases = map[StatusCode]string{
	OK:                    "OK",
	BAD_REQUEST:           "Bad Request",
	INTERNAL_SERVER_ERROR: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	rp := ""
	if v, ok := reasonPhrases[statusCode]; ok {
		rp = v
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, rp)
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) header.Header {
	h := header.Header{}

	h.Set("Content-Length", contentLen)
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeader(w io.Writer, header header.Header) error {
	formattedHeaders := ""
	for key, value := range header {
		formattedHeaders += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	_, err := w.Write([]byte(formattedHeaders))
	return err
}
