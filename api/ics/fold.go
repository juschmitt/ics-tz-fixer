package ics

import "strings"

func foldLine(line string) []string {
	runes := []rune(line)
	if len(runes) <= 75 {
		return []string{line}
	}

	chunks := []string{string(runes[:75])}
	remaining := runes[75:]
	for len(remaining) > 0 {
		chunkSize := 74
		if len(remaining) < chunkSize {
			chunkSize = len(remaining)
		}
		chunks = append(chunks, " "+string(remaining[:chunkSize]))
		remaining = remaining[chunkSize:]
	}

	return chunks
}

func unfoldLines(content string) []string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(content, "\n")
	unfolded := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if len(unfolded) == 0 {
				unfolded = append(unfolded, strings.TrimLeft(line, " \t"))
				continue
			}
			unfolded[len(unfolded)-1] += strings.TrimLeft(line, " \t")
			continue
		}
		unfolded = append(unfolded, line)
	}
	return unfolded
}
