package router

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type compressedResponseWriter struct {
	http.ResponseWriter
	writer io.Writer
}

func (w *compressedResponseWriter) Write(b []byte) (int, error) {
	return w.writer.Write(b)
}

func compressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encoding := r.Header.Get("Accept-Encoding")
		contentType := w.Header().Get("Content-Type")

		if strings.Contains(encoding, "gzip") && (contentType == "application/json" || contentType == "text/html") {
			w.Header().Set("Content-Encoding", "gzip")
			gzWriter := gzip.NewWriter(w)
			defer gzWriter.Close()
			w = &compressedResponseWriter{ResponseWriter: w, writer: gzWriter}
		}

		if strings.Contains(encoding, "deflate") && (contentType == "application/json" || contentType == "text/html") {
			w.Header().Set("Content-Encoding", "deflate")
			flWriter, _ := flate.NewWriter(w, flate.DefaultCompression)
			defer flWriter.Close()
			w = &compressedResponseWriter{ResponseWriter: w, writer: flWriter}
		}

		next.ServeHTTP(w, r)
	})
}

func decompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentEncoding := r.Header.Get("Content-Encoding")

		switch contentEncoding {
		case "gzip":
			gzReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "could not decompress gzip body", http.StatusBadRequest)
				return
			}
			defer gzReader.Close()
			r.Body = io.NopCloser(gzReader)

		case "deflate":
			flReader := flate.NewReader(r.Body)
			defer flReader.Close()
			r.Body = io.NopCloser(flReader)

		case "":

		default:
			http.Error(w, "unsupported content encoding", http.StatusUnsupportedMediaType)
			return
		}

		next.ServeHTTP(w, r)
	})
}
