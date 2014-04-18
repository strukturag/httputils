// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
	"strings"
)

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// MakeGzipHandler wraps handler such that its output will be compressed
// according to what the client supports.
func MakeGzipHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encodings := strings.Split(r.Header.Get("Accept-Encoding"), ",")
		var w_compressed io.WriteCloser
		var err error
		for _, encoding := range encodings {
			encoding = strings.TrimSpace(encoding)
			switch encoding {
			case "gzip":
				w_compressed, err = gzip.NewWriterLevel(w, gzip.BestSpeed)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer w_compressed.Close()
				w.Header().Set("Content-Encoding", "gzip")
			case "deflate":
				w_compressed, err = zlib.NewWriterLevel(w, zlib.BestSpeed)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer w_compressed.Close()
				w.Header().Set("Content-Encoding", "deflate")
			}
			if w_compressed != nil {
				break
			}
		}
		if w_compressed == nil {
			handler(w, r)
			return
		}
		w.Header().Set("Vary", "Accept-Encoding")
		handler(gzipResponseWriter{Writer: w_compressed, ResponseWriter: w}, r)

	}
}
