package log

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type StructuredLogger struct {
	logger *logrus.Logger
}

func NewStructuredLogger(logger *logrus.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{logger})
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &StructuredLoggerEntry{logger: logrus.NewEntry(l.logger)}
	logFields := logrus.Fields{}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	logFields["http_proto"] = r.Proto
	logFields["http_method"] = r.Method

	logFields["remote_addr"] = r.RemoteAddr
	logFields["user_agent"] = r.UserAgent()

	logFields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	entry.logger = entry.logger.WithFields(logFields)

	return entry
}

type StructuredLoggerEntry struct {
	logger logrus.FieldLogger
}

func (l *StructuredLoggerEntry) Write(status, bytes int,header http.Header, elapsed time.Duration, extra interface{}) {
	l.logger = l.logger.WithFields(logrus.Fields{
		"resp_status":       status,
		"resp_bytes_length": bytes,
		"resp_elapsed":      elapsed.Seconds(),
	})

	l.logger.Infoln("request complete")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	l.logger = l.logger.WithFields(logrus.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	})
}

// Helper methods used by the application to get the request-scoped
// logger entry and set additional fields between handlers.
//
// This is a useful pattern to use to set state on the entry as it
// passes through the handler chain, which at any point can be logged
// with a call to .Print(), .Info(), etc.

func GetLogEntry(r *http.Request) logrus.FieldLogger {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	return entry.logger
}

func LogEntrySetField(r *http.Request, key string, value interface{}) {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	entry.logger = entry.logger.WithField(key, value)
	middleware.WithLogEntry(r, entry)
}

func LogEntrySetFields(r *http.Request, fields map[string]interface{}) {
	entry := middleware.GetLogEntry(r).(*StructuredLoggerEntry)
	entry.logger = entry.logger.WithFields(fields)
	middleware.WithLogEntry(r, entry)
}
