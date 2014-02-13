// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
    "net/http"
    "os"
)

// ServeFile responds to w with the contents of path within fs.
func ServeFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem, path string) {

    // Open file handle.
    f, err := fs.Open(path)
    if err != nil {
        http.Error(w, "404 Not Found", 404)
        return
    }
    defer f.Close()

    // Make sure path exists.
    fileinfo, err1 := f.Stat()
    if err1 != nil {
        http.Error(w, "404 Not Found", 404)
        return
    }

    // Reject directory requests.
    if fileinfo.IsDir() {
        http.Error(w, "403 Forbidden", 403)
        return
    }

    http.ServeContent(w, r, fileinfo.Name(), fileinfo.ModTime(), f)

}

// HasFilePath returns true if path is openable and stat-able, otherwise false.
func HasFilePath(path string) bool {

    f, err := os.Open(path)
    if err != nil {
        return false
    }
    defer f.Close()
    _, err = f.Stat()
    if err != nil {
        return false
    }
    return true

}

// HasDirPath returns true if path is openable, stat-able, and a directory.
func HasDirPath(path string) bool {

    f, err := os.Open(path)
    if err != nil {
        return false
    }
    defer f.Close()
    fileinfo, err := f.Stat()
    if err != nil {
        return false
    }
    if !fileinfo.IsDir() {
        return false
    }
    return true

}
