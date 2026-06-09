package util

import "strings"

func ParseASCIIToString(part []byte) string {
	var asciiString string
	for _, code := range part {
		asciiString += string(rune(code))
	}
	return strings.TrimSpace(asciiString)
}
