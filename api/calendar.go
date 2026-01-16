package handler

import (
	"net/http"

	"github.com/juschmitt/ics-tz-fixer/api/http/calendar"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	calendar.Handler(w, r)
}
