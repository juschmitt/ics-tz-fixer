package calendar

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestHandlerRejectsDisallowedHost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/calendar?url=https://example.com/feed.ics", nil)
	recorder := httptest.NewRecorder()

	Handler(recorder, req)

	if recorder.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Result().StatusCode)
	}
}

func TestHandlerRejectsNonCalendarContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"))
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}
	allowHost(t, parsed.Hostname())

	requestURL := "/api/calendar?url=" + url.QueryEscape(server.URL)
	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	recorder := httptest.NewRecorder()

	Handler(recorder, req)

	if recorder.Result().StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("expected status 415, got %d", recorder.Result().StatusCode)
	}
}

func TestHandlerRejectsInvalidCalendarPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte("NOT-A-CALENDAR\r\n"))
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}
	allowHost(t, parsed.Hostname())

	requestURL := "/api/calendar?url=" + url.QueryEscape(server.URL)
	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	recorder := httptest.NewRecorder()

	Handler(recorder, req)

	if recorder.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", recorder.Result().StatusCode)
	}
}

func TestHandlerRejectsLargeCalendar(t *testing.T) {
	largePayload := "BEGIN:VCALENDAR\r\n" + strings.Repeat("A", maxICSBytes)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/calendar")
		_, _ = w.Write([]byte(largePayload))
	}))
	t.Cleanup(server.Close)

	parsed, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}
	allowHost(t, parsed.Hostname())

	requestURL := "/api/calendar?url=" + url.QueryEscape(server.URL)
	req := httptest.NewRequest(http.MethodGet, requestURL, nil)
	recorder := httptest.NewRecorder()

	Handler(recorder, req)

	if recorder.Result().StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status 413, got %d", recorder.Result().StatusCode)
	}
}

func TestIsValidTZHint(t *testing.T) {
	cases := []struct {
		value    string
		expected bool
	}{
		{value: "Europe/Berlin", expected: true},
		{value: "America/New_York", expected: true},
		{value: "EuropeBerlin", expected: false},
		{value: "Europe/", expected: false},
		{value: "Europe /Berlin", expected: false},
		{value: "Europe/Berlin/Extra", expected: false},
		{value: "Europe/Berlin+Test", expected: true},
	}

	for _, testCase := range cases {
		if isValidTZHint(testCase.value) != testCase.expected {
			t.Fatalf("tz_hint validation mismatch for %q", testCase.value)
		}
	}
}

func allowHost(t *testing.T, host string) {
	t.Helper()
	allowedHosts[host] = struct{}{}
	t.Cleanup(func() {
		delete(allowedHosts, host)
	})
}
