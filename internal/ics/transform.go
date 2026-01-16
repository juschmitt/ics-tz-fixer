package ics

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strings"
)

type tzRule struct {
	standardOffsetFrom string
	standardOffsetTo   string
	standardRRule      string
	daylightOffsetFrom string
	daylightOffsetTo   string
	daylightRRule      string
}

func (rule tzRule) signature() string {
	return fmt.Sprintf(
		"STD:%s>%s;%s|DST:%s>%s;%s",
		rule.standardOffsetFrom,
		rule.standardOffsetTo,
		rule.standardRRule,
		rule.daylightOffsetFrom,
		rule.daylightOffsetTo,
		rule.daylightRRule,
	)
}

func Transform(reader io.Reader, tzHint string) (io.Reader, error) {
	tzHint = strings.TrimSpace(tzHint)
	pipeReader, pipeWriter := io.Pipe()

	go func() {
		err := transformStream(reader, pipeWriter, tzHint)
		_ = pipeWriter.CloseWithError(err)
	}()

	return pipeReader, nil
}

func transformStream(reader io.Reader, writer io.Writer, tzHint string) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 1024), 1024*1024)

	signatureMapping := knownSignatureMapping()
	mapping := map[string]string{}

	var (
		pendingLine string
		inTimezone  bool
		section     string
		currentRule tzRule
		currentTZID string
	)

	flushLine := func(line string) error {
		for _, folded := range foldLine(line) {
			if _, err := io.WriteString(writer, folded+"\r\n"); err != nil {
				return err
			}
		}
		return nil
	}

	processLine := func(line string) error {
		if line == "BEGIN:VTIMEZONE" {
			inTimezone = true
			section = ""
			currentRule = tzRule{}
			currentTZID = ""
			return nil
		}

		if inTimezone {
			switch line {
			case "BEGIN:STANDARD":
				section = "STANDARD"
				return nil
			case "END:STANDARD":
				section = ""
				return nil
			case "BEGIN:DAYLIGHT":
				section = "DAYLIGHT"
				return nil
			case "END:DAYLIGHT":
				section = ""
				return nil
			case "END:VTIMEZONE":
				if currentTZID != "" {
					applyMapping(mapping, signatureMapping, currentTZID, currentRule, tzHint)
				}
				inTimezone = false
				return nil
			}

			if strings.HasPrefix(line, "TZID:") {
				currentTZID = strings.TrimSpace(strings.TrimPrefix(line, "TZID:"))
				return nil
			}

			if section == "" {
				return nil
			}

			key, value, ok := splitLine(line)
			if !ok {
				return nil
			}
			value = strings.TrimSpace(value)

			switch key {
			case "TZOFFSETFROM":
				if section == "STANDARD" {
					currentRule.standardOffsetFrom = value
				} else {
					currentRule.daylightOffsetFrom = value
				}
			case "TZOFFSETTO":
				if section == "STANDARD" {
					currentRule.standardOffsetTo = value
				} else {
					currentRule.daylightOffsetTo = value
				}
			case "RRULE":
				if section == "STANDARD" {
					currentRule.standardRRule = normalizeRRule(value)
				} else {
					currentRule.daylightRRule = normalizeRRule(value)
				}
			}

			return nil
		}

		return flushLine(updateLine(line, mapping))
	}

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if len(line) > 0 && (line[0] == ' ' || line[0] == '\t') {
			if pendingLine == "" {
				pendingLine = strings.TrimLeft(line, " \t")
				continue
			}
			pendingLine += strings.TrimLeft(line, " \t")
			continue
		}

		if pendingLine != "" {
			if err := processLine(pendingLine); err != nil {
				return err
			}
		}
		pendingLine = line
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if pendingLine != "" {
		if err := processLine(pendingLine); err != nil {
			return err
		}
	}

	return nil
}

func applyMapping(mapping, signatureMapping map[string]string, tzid string, rule tzRule, tzHint string) {
	if tzHint != "" {
		mapping[tzid] = tzHint
		return
	}

	if iana, ok := signatureMapping[rule.signature()]; ok {
		mapping[tzid] = iana
	}
}

func normalizeRRule(rule string) string {
	if rule == "" {
		return ""
	}
	parts := strings.Split(rule, ";")
	cleaned := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, strings.ToUpper(trimmed))
		}
	}
	sort.Strings(cleaned)
	return strings.Join(cleaned, ";")
}

func updateLine(line string, mapping map[string]string) string {
	key, value, ok := splitLine(line)
	if !ok {
		return line
	}

	if key == "TZID" {
		if updated, found := resolveTZID(value, mapping); found {
			return "TZID:" + updated
		}
		return line
	}

	if strings.Contains(key, "TZID=") {
		key = replaceTZIDParam(key, mapping)
		return key + ":" + value
	}

	return line
}

func replaceTZIDParam(key string, mapping map[string]string) string {
	parts := strings.Split(key, ";")
	if len(parts) == 1 {
		return key
	}

	for index := 1; index < len(parts); index++ {
		part := parts[index]
		if !strings.HasPrefix(part, "TZID=") {
			continue
		}

		value := strings.TrimPrefix(part, "TZID=")
		if updated, found := resolveTZID(value, mapping); found {
			parts[index] = "TZID=" + updated
		}
	}

	return strings.Join(parts, ";")
}

func resolveTZID(tzid string, mapping map[string]string) (string, bool) {
	if updated, found := mapping[tzid]; found {
		return updated, true
	}

	if strings.HasSuffix(tzid, " Standard Time") {
		trimmed := strings.TrimSuffix(tzid, " Standard Time")
		if updated, found := mapping[trimmed]; found {
			return updated, true
		}
	}

	return "", false
}
