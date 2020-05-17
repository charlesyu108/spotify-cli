package utils

import (
	"fmt"
	"strings"
)

// Check for errors
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

// Sprintf except only for strings and strings are trimmed before formatting.
func FormatString(template string, strs ...string) string {
	trimmed := make([]interface{}, len(strs))
	for idx, s := range strs {
		trimmed[idx] = strings.TrimSpace(s)
	}
	return fmt.Sprintf(template, trimmed...)
}
