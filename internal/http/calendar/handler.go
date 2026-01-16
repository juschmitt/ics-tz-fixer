package calendar

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/juschmitt/ics-tz-fixer/internal/ics"
)

var allowedHosts = map[string]struct{}{
	"outlook.office365.com": {},
}

const maxICSBytes = 5 * 1024 * 1024

var tzHintPattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]+/[A-Za-z0-9_+.-]+$`)

func Handler(w http.ResponseWriter, r *http.Request) {
	sourceURL := r.URL.Query().Get("url")
	if sourceURL == "" {
		http.Error(w, "missing url query parameter", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(sourceURL)
	if err != nil || (parsedURL.Scheme != "https" && parsedURL.Scheme != "http") {
		http.Error(w, "invalid url query parameter", http.StatusBadRequest)
		return
	}

	if !isAllowedHost(parsedURL.Hostname()) {
		http.Error(w, "source host is not allowed", http.StatusForbidden)
		return
	}

	tzHint := strings.TrimSpace(r.URL.Query().Get("tz_hint"))
	if tzHint != "" && !isValidTZHint(tzHint) {
		http.Error(w, "invalid tz_hint query parameter", http.StatusBadRequest)
		return
	}

	request, err := http.NewRequestWithContext(r.Context(), http.MethodGet, sourceURL, nil)
	if err != nil {
		http.Error(w, "failed to build request", http.StatusBadRequest)
		return
	}

	client := &http.Client{Timeout: 30 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		http.Error(w, "failed to fetch source calendar", http.StatusBadGateway)
		return
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		http.Error(w, "source calendar returned non-2xx status", http.StatusBadGateway)
		return
	}

	if !isCalendarContentType(response.Header.Get("Content-Type")) {
		http.Error(w, "source calendar is not text/calendar", http.StatusUnsupportedMediaType)
		return
	}

	payload, err := io.ReadAll(io.LimitReader(response.Body, maxICSBytes+1))
	if err != nil {
		http.Error(w, "failed to read source calendar", http.StatusBadGateway)
		return
	}
	if len(payload) > maxICSBytes {
		http.Error(w, "source calendar exceeds size limit", http.StatusRequestEntityTooLarge)
		return
	}
	if err := validateCalendarPayload(payload); err != nil {
		http.Error(w, "source calendar is not valid", http.StatusBadRequest)
		return
	}

	transformed, err := ics.Transform(bytes.NewReader(payload), tzHint)
	if err != nil {
		http.Error(w, "failed to transform calendar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, transformed)
}

func isAllowedHost(host string) bool {
	_, ok := allowedHosts[strings.ToLower(host)]
	return ok
}

func isValidTZHint(tzHint string) bool {
	return tzHintPattern.MatchString(tzHint)
}

func isCalendarContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return strings.Contains(contentType, "text/calendar")
}

func validateCalendarPayload(payload []byte) error {
	reader := bufio.NewScanner(bytes.NewReader(payload))
	for reader.Scan() {
		line := strings.TrimSpace(strings.TrimSuffix(reader.Text(), "\r"))
		if line == "" {
			continue
		}
		if line != "BEGIN:VCALENDAR" {
			return fmt.Errorf("missing BEGIN:VCALENDAR")
		}
		return nil
	}
	if err := reader.Err(); err != nil {
		return err
	}
	return fmt.Errorf("empty calendar")
}
