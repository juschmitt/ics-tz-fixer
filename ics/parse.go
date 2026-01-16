package ics

import "strings"

func splitLine(line string) (string, string, bool) {
	colon := strings.IndexByte(line, ':')
	if colon == -1 {
		return "", "", false
	}
	return line[:colon], line[colon+1:], true
}
