package server

import (
	"bytes"
	"io"
	"net/http"
	"time"
)

func (h *Handlers) RequestLogger(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := NewResponseLoggerWriter(w)

		var buf bytes.Buffer

		tee := io.TeeReader(r.Body, &buf)
		body, err := io.ReadAll(tee)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			h.log.Errorf("request logger read body err: %w", err)
			return
		}
		r.Body = io.NopCloser(&buf)

		start := time.Now()
		hr.ServeHTTP(rw, r)
		duration := time.Since(start)

		h.log.Infof("HTTP request method: %s, header: %v, body: %s, url: %s, duration: %s, statusCode: %d, responseSize: %d",
			r.Method, r.Header, string(body), r.RequestURI, duration, rw.responseData.status, rw.responseData.size,
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
