package main

import (
	"net/http"

	"github.com/HeavyHorst/sunoKB/pkg/log"
)

func logAndHTTPError(w http.ResponseWriter, r *http.Request, code int, httptext string, err error) {
	logger := log.GetLogEntry(r)
	logger.Error(err)
	http.Error(w, httptext, code)
}
