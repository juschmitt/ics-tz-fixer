package handler

import (
	"net/http"

	"github.com/juschmitt/ics-tz-fixer/internal/http/calendar"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	calendar.Handler(w, r)
}
