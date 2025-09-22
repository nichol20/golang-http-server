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

	fmt.Printf("Field line: %s\nField name: %s\nField value: %s\n", fieldLine, fieldName, fieldValue)

	pattern := `^[A-Za-z0-9!#$%&'*+\-.\^_` + "`" + `|~]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(fieldName) {
		return 0, false, fmt.Errorf("invalid field name")
	}

	consumed := crlfIdx + len(CRLF)
	loweredFieldName := strings.ToLower(fieldName)

	if _, exists := h[loweredFieldName]; exists {
		h[loweredFieldName] = h[loweredFieldName] + ", " + strings.TrimSpace(fieldValue)
		return consumed, false, nil
	}

	h[loweredFieldName] = strings.TrimSpace(fieldValue)
	return consumed, false, nil
}

func (h Header) Get(key string) string {
	value, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return value
}

func (h Header) Set(key, value string) {
	h[key] = strings.ToLower(value)
}
