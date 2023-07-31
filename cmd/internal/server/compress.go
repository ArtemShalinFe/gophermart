package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const compressedTypes = "application/json,text/html"

func CompressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(compressedTypes, contentType) {
			h.ServeHTTP(w, r)
		}
		origWriter := w

		acceptEncodings := r.Header.Values("Accept-Encoding")
		for _, acceptEncoding := range acceptEncodings {
			if strings.Contains(acceptEncoding, "gzip") {
				gzipWriter := NewGzipWriter(w)
				origWriter = gzipWriter
				defer gzipWriter.Close()
			}
		}
		contentEncodings := r.Header.Values("Content-Encoding")

		for _, contentEncoding := range contentEncodings {
			if strings.Contains(contentEncoding, "gzip") {
				compressReader, err := NewGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = compressReader
				defer compressReader.Close()
			}
		}

		h.ServeHTTP(origWriter, r)
	})
}

type gzipWriter struct {
	http.ResponseWriter
	zipW *gzip.Writer
}

func NewGzipWriter(w http.ResponseWriter) *gzipWriter {
	return &gzipWriter{
		ResponseWriter: w,
		zipW:           gzip.NewWriter(w),
	}
}

func (c *gzipWriter) Write(p []byte) (int, error) {
	return c.zipW.Write(p)
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.ResponseWriter.Header().Set("Content-Encoding", "gzip")
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	return c.zipW.Close()
}

type gzipReader struct {
	r    io.ReadCloser
	zipR *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &gzipReader{
		r:    r,
		zipR: zipR,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	return c.zipR.Read(p)
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zipR.Close()
}
