// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
    "fmt"
    "net/http"
    "path"
    "strings"
    "time"
)

type fileStaticHandler struct {
    root http.FileSystem
}

// FileStaticServer returns a handler that serves HTTP requests
// with the contents of the file system rooted at root and
// far future caching headers set.
//
// To use the operating system's file system implementation,
// use http.Dir:
//
//     http.Handle("/", http.FileStaticServer(http.Dir("/tmp")))
func FileStaticServer(root http.FileSystem) http.Handler {
    return &fileStaticHandler{root}
}

func (f *fileStaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    upath := r.URL.Path
    if !strings.HasPrefix(upath, "/") {
        upath = "/" + upath
        r.URL.Path = upath
    }

    handler := func(w http.ResponseWriter, r *http.Request) {

        upath = path.Clean(upath)
        parts := strings.Split(upath, "/")
        if len(parts) > 3 && strings.HasPrefix(parts[2], "ver=") {
            // Filter version from upath
            upath = fmt.Sprintf("%s/%s", strings.Join(parts[:2], "/"), strings.Join(parts[3:], "/"))
            // Add far futore expire header
            w.Header().Set("Expires", (time.Now().UTC().AddDate(1, 0, 0).Format(http.TimeFormat)))
            w.Header().Set("Cache-Control", "public, max-age=31536000")
            w.Header().Set("X-Content-Type-Options", "nosniff")
        }

        ServeFile(w, r, f.root, upath)

    }

    MakeGzipHandler(handler)(w, r)

}
