package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressReader struct {
	io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &compressReader{r, gz}, nil
}

func (cr *compressReader) Read(p []byte) (n int, err error) {
	return cr.zr.Read(p)
}

func (cr *compressReader) Close() error {
	return cr.zr.Close()
}

func RequestDecompressor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") && r.ContentLength > 0 {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				http.Error(w, "Failed to decode gzip body", http.StatusBadRequest)
				return
			}
			defer cr.Close()
			r.Body = cr
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length")
		}
		next.ServeHTTP(w, r)
	})
}
