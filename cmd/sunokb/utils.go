package main

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/HeavyHorst/sunoKB/pkg/log"
	"github.com/oklog/ulid"
)

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

func getULID() string {
	entropy := randPool.Get().(*rand.Rand)
	defer randPool.Put(entropy)
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}

func logAndHTTPError(w http.ResponseWriter, r *http.Request, code int, httptext string, err error) {
	logger := log.GetLogEntry(r)
	logger.Error(err)
	http.Error(w, httptext, code)
}
