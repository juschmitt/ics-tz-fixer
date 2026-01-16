package ics

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestTransformRewritesKnownTimezone(t *testing.T) {
	reader := readFixture(t, "basic.ics")

	out, err := Transform(reader, "")
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}

	output := readAll(t, out)
	if !strings.Contains(output, "DTSTART;TZID=Europe/Berlin:20250115T105000") {
		t.Fatalf("expected DTSTART to use Europe/Berlin")
	}
	if strings.Contains(output, "W. Europe Standard Time") {
		t.Fatalf("expected original TZID to be removed")
	}
	if strings.Contains(output, "BEGIN:VTIMEZONE") {
		t.Fatalf("expected VTIMEZONE block to be stripped")
	}
}

func TestTransformTZHintOverridesMapping(t *testing.T) {
	reader := readFixture(t, "hint.ics")

	out, err := Transform(reader, "Europe/Paris")
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}

	output := readAll(t, out)
	if !strings.Contains(output, "DTSTART;TZID=Europe/Paris:20250115T114500") {
		t.Fatalf("expected tz_hint to override timezone mapping")
	}
	if strings.Contains(output, "Central Europe Standard Time") {
		t.Fatalf("expected original TZID to be removed")
	}
}

func TestTransformLeavesUnknownTimezone(t *testing.T) {
	reader := readFixture(t, "unknown.ics")

	out, err := Transform(reader, "")
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}

	output := readAll(t, out)
	if strings.Contains(output, "TZID:Custom Zone") {
		t.Fatalf("expected VTIMEZONE block to be stripped")
	}
	if !strings.Contains(output, "DTSTART;TZID=Custom Zone:20250115T090000") {
		t.Fatalf("expected DTSTART to remain unchanged")
	}
}

func TestTransformUnfoldsAndRefoldsLines(t *testing.T) {
	reader := readFixture(t, "folded.ics")

	out, err := Transform(reader, "")
	if err != nil {
		t.Fatalf("Transform returned error: %v", err)
	}

	output := readAll(t, out)
	unfolded := strings.Join(unfoldLines(output), "\n")
	if !strings.Contains(unfolded, "SUMMARY:"+strings.Repeat("A", 90)) {
		t.Fatalf("expected folded SUMMARY to be preserved")
	}
}

func readFixture(t *testing.T, name string) io.Reader {
	fixturePath := "testdata/" + name
	content, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", fixturePath, err)
	}
	return strings.NewReader(string(content))
}

func readAll(t *testing.T, reader io.Reader) string {
	content, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}
	return string(content)
}
