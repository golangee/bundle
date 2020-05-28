package bundle

import (
	"strings"
	"unicode"
)

// slashToCamelCase makes from /my/path a string like MyPath
func slashToCamelCase(str string) string {
	sb := &strings.Builder{}
	nextUp := true
	for _, r := range str {
		if r == '/' {
			nextUp = true
			continue
		}

		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			nextUp = true
			continue
		}

		if nextUp {
			sb.WriteRune(unicode.ToUpper(r))
			nextUp = false
		} else {
			sb.WriteRune(r)
		}

	}
	return sb.String()
}
