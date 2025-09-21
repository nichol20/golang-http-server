package headers

import (
	"fmt"
	"strings"
)

const CRLF = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
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

	if i := strings.Index(fieldName, " "); i != -1 {
		return 0, false, fmt.Errorf("invalid field name")
	}

	h[fieldName] = strings.TrimSpace(fieldValue)
	return crlfIdx + len(CRLF), false, nil
}
