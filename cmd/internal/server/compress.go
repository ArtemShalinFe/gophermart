package server

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const compressedTypes = "application/json,text/html"
const gzipType = "gzip"
const contentEncoding = "Content-Encoding"

func (h *Handlers) CompressMiddleware(hr http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get(contentType)
		if !strings.Contains(compressedTypes, contentType) {
			hr.ServeHTTP(w, r)
		}
		origWriter := w

		acceptEncodings := r.Header.Values("Accept-Encoding")
		for _, acceptEncoding := range acceptEncodings {
			if strings.Contains(acceptEncoding, gzipType) {
				gzipWriter := NewGzipWriter(w)
				origWriter = gzipWriter
				defer func() {
					if err := gzipWriter.Close(); err != nil {
						h.log.Errorf("close gzip writer failed err: %w", err)
					}
				}()
			}
		}
		contentEncodings := r.Header.Values(contentEncoding)

		for _, contentEncoding := range contentEncodings {
			if strings.Contains(contentEncoding, gzipType) {
				compressReader, err := NewGzipReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = compressReader
				defer func() {
					if err := compressReader.Close(); err != nil {
						h.log.Errorf("close compress reader failed err: %w", err)
					}
				}()
			}
		}
		hr.ServeHTTP(origWriter, r)
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
	n, err := c.zipW.Write(p)
	if err != nil {
		return 0, fmt.Errorf("gzip writer write failed err: %w", err)
	}
	return n, nil
}

func (c *gzipWriter) WriteHeader(statusCode int) {
	if statusCode < http.StatusMultipleChoices {
		c.ResponseWriter.Header().Set(contentEncoding, gzipType)
	}
	c.ResponseWriter.WriteHeader(statusCode)
}

func (c *gzipWriter) Close() error {
	if err := c.zipW.Close(); err != nil {
		return fmt.Errorf("gzip writer close failed err: %w", err)
	}
	return nil
}

type gzipReader struct {
	r    io.ReadCloser
	zipR *gzip.Reader
}

func NewGzipReader(r io.ReadCloser) (*gzipReader, error) {
	zipR, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("init gzip reader failed err: %w", err)
	}

	return &gzipReader{
		r:    r,
		zipR: zipR,
	}, nil
}

func (c gzipReader) Read(p []byte) (n int, err error) {
	n, err = c.zipR.Read(p)
	if err != nil {
		return 0, fmt.Errorf("gzip reader read failed err: %w", err)
	}
	return n, nil
}

func (c *gzipReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("reader close failed err: %w", err)
	}
	if err := c.zipR.Close(); err != nil {
		return fmt.Errorf("gzip reader close failed err: %w", err)
	}
	return nil
}
