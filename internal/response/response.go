package response

import (
	"fmt"
	"io"

	"github.com/nichol20/http-server/internal/header"
)

type StatusCode uint16

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

var reasonPhrases = map[StatusCode]string{
	StatusOK:                  "OK",
	StatusBadRequest:          "Bad Request",
	StatusInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode int16) error {
	rp := ""
	if v, ok := reasonPhrases[StatusCode(statusCode)]; ok {
		rp = v
	}

	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, rp)
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) header.Header {
	h := header.Header{}

	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	return h
}

func WriteHeader(w io.Writer, header header.Header) error {
	b := []byte{}
	for key, value := range header {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	}
	b = fmt.Append(b, "\r\n")
	_, err := w.Write(b)
	return err
}
