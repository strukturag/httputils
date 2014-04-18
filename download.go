// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"net/http"
	"path"
	"strings"
)

type fileDownloadHandler struct {
	root http.FileSystem
}

// FileDownloadServer returns a handler that serves HTTP requests
// with the contents of the file system rooted at root and
// file download headers set.
//
// To use the operating system's file system implementation,
// use http.Dir:
//
//     http.Handle("/", http.FileStaticServer(http.Dir("/tmp")))
func FileDownloadServer(root http.FileSystem) http.Handler {
	return &fileDownloadHandler{root}
}

func (f *fileDownloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	upath := r.URL.Path
	if !strings.HasPrefix(upath, "/") {
		upath = "/" + upath
		r.URL.Path = upath
	}

	handler := func(w http.ResponseWriter, r *http.Request) {

		upath = path.Clean(upath)
		_, fn := path.Split(upath)
		w.Header().Set("Content-Disposition", "attachment;filename=\""+fn+"\"")

		ServeFile(w, r, f.root, upath)

	}

	handler(w, r)

}
