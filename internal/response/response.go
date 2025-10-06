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

type Writer struct {
	ioWriter io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		ioWriter: w,
	}
}

func (w *Writer) WriteStatusLine(statusCode int16) error {
	rp := ""
	if v, ok := reasonPhrases[StatusCode(statusCode)]; ok {
		rp = v
	}
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, rp)
	_, err := w.ioWriter.Write([]byte(statusLine))
	return err
}

func (w *Writer) WriteHeader(header header.Header) error {
	b := []byte{}
	for key, value := range header {
		b = fmt.Appendf(b, "%s: %s\r\n", key, value)
	}
	b = fmt.Append(b, "\r\n")
	_, err := w.ioWriter.Write(b)
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	return w.ioWriter.Write(p)
}

func (w *Writer) WriteRespose(statusCode int16, header header.Header, message []byte) error {
	err := w.WriteStatusLine(statusCode)
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}
	err = w.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}
	_, err = w.WriteBody(message)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	return nil
}

func GetDefaultHeaders(contentLen int) header.Header {
	h := header.Header{}
	h.Set("Content-Length", fmt.Sprintf("%d", contentLen))
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")
	return h
}
