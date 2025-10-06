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

func (h Header) parseSingle(data []byte) (n int, done bool, err error) {
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

func (h Header) Parse(data []byte) (n int, done bool, err error) {
	total := 0

	for {
		n, done, err := h.parseSingle(data[total:])
		if n <= 0 {
			return total, done, err
		}
		if err != nil {
			return 0, done, err
		}
		total += n
		if done {
			return total, done, err
		}
	}
}

func (h Header) Get(key string) string {
	value, ok := h[strings.ToLower(key)]
	if !ok {
		return ""
	}
	return value
}

func (h Header) Set(key string, value string) {
	loweredKey := strings.ToLower(key)
	if _, exists := h[loweredKey]; exists {
		// space between commas is optional
		h[loweredKey] = fmt.Sprintf("%s, %s", h[loweredKey], value)
	} else {
		h[loweredKey] = fmt.Sprintf("%v", value)
	}
}

func (h Header) Del(key string) {
	loweredKey := strings.ToLower(key)
	delete(h, loweredKey)
}

func (h Header) Replace(key string, value string) {
	h.Del(key)
	h.Set(key, value)
}
