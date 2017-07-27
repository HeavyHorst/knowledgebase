package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/HeavyHorst/knowledgebase/pkg/log"
	"github.com/HeavyHorst/knowledgebase/pkg/models"
)

func isLoggedIn(ctx context.Context) bool {
	_, ok := ctx.Value(contextKeyCurrentUser).(models.User)
	return ok
}

func logAndHTTPError(w http.ResponseWriter, r *http.Request, code int, httptext string, err error) {
	logger := log.GetLogEntry(r)
	logger.Error(err)
	http.Error(w, httptext, code)
}

func writeJSON(w http.ResponseWriter, r *http.Request, value interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		logAndHTTPError(w, r, 500, err.Error(), err)
	}
}
