package middleware

import (
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"
)

type compressWriter struct {
	http.ResponseWriter
	zw             *gzip.Writer
	shouldCompress bool
}

func newCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		ResponseWriter: w,
		zw:             gzip.NewWriter(w),
		shouldCompress: false,
	}
}

func (cw *compressWriter) Header() http.Header {
	return cw.ResponseWriter.Header()
}

func (cw *compressWriter) Write(p []byte) (int, error) {
	if cw.shouldCompress {
		return cw.zw.Write(p)
	}
	return cw.ResponseWriter.Write(p)
}

func canCompress(contentType string) bool {
	return strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html")
}

func (cw *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		contentType := cw.Header().Get("Content-Type")
		if canCompress(contentType) {
			cw.shouldCompress = true
			cw.Header().Set("Content-Encoding", "gzip")
			cw.Header().Del("Content-Length")
		}
	}
	cw.ResponseWriter.WriteHeader(statusCode)
}
func (cw *compressWriter) Close() error {
	if cw.shouldCompress {
		return cw.zw.Close()
	}
	return nil
}

func ResponseCompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			cw := newCompressWriter(w)
			defer cw.Close()
			w = cw
		}
		next.ServeHTTP(w, r)
	})
}
