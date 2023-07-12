package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (h *Handlers) RequestLogger(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rw := NewResponseLoggerWriter(w)

		var buf bytes.Buffer

		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.log.Error("request logger read body ", zap.Error(err))
			return
		}
		r.Body = io.NopCloser(&buf)

		start := time.Now()
		hr.ServeHTTP(rw, r)
		duration := time.Since(start)

		h.log.Info(
			"HTTP request",
			zap.String("method", r.Method),
			// TODO
			// Так не пойдет, теряется смысл "не сахарного логгера".
			// Придумать что-нибудь с fmt.Sprintf и заголовками.
			zap.String("headers", fmt.Sprintf("%+v", r.Header)),
			zap.String("body", string(body)),
			zap.String("url", r.RequestURI),
			zap.Duration("duration", duration),
			zap.Int("status", rw.responseData.status),
			zap.Int("size", rw.responseData.size),
		)

	})
}

type responseData struct {
	status int
	size   int
}

type ResponseLoggerWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func NewResponseLoggerWriter(w http.ResponseWriter) *ResponseLoggerWriter {
	return &ResponseLoggerWriter{
		ResponseWriter: w,
		responseData:   &responseData{},
	}
}

func (r *ResponseLoggerWriter) Write(b []byte) (int, error) {

	size, err := r.ResponseWriter.Write(b)
	if err != nil {
		return 0, err
	}

	r.responseData.size += size

	return size, nil

}

func (r *ResponseLoggerWriter) WriteHeader(statusCode int) {

	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode

}
