package header

import (
	"fmt"
	"regexp"
	"strings"
)

const CRLF = "\r\n"

type Header map[string]string

func NewHeader() Header {
	return Header{}
}

func (h Header) Parse(data []byte) (n int, done bool, err error) {
	dataStr := string(data)
	crlfIdx := strings.Index(dataStr, CRLF)
	if crlfIdx == -1 {
		return 0, false, nil
	}
	if crlfIdx == 0 {
		return len(CRLF), true, nil
	}

	fieldLine := dataStr[:crlfIdx]
	fieldLine = strings.Trim(fieldLine, " ")

	colIdx := strings.Index(fieldLine, ":")
	if colIdx == -1 {
		return 0, false, fmt.Errorf("malformed field line")
	}
	fieldName, fieldValue := fieldLine[:colIdx], fieldLine[colIdx+1:]

	pattern := `^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(fieldName) {
		return 0, false, fmt.Errorf("invalid field name")
	}

	h.Set(fieldName, strings.TrimSpace(fieldValue))
	return crlfIdx + len(CRLF), false, nil
}

func (h Header) Get(key string) string {
	value, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return value
}

func (h Header) Set(key, value string) {
	loweredKey := strings.ToLower(key)
	if _, exists := h[loweredKey]; exists {
		h[loweredKey] = fmt.Sprintf("%s, %s", h[loweredKey], value)
	} else {
		h[loweredKey] = value
	}
}
