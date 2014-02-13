// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeWelcomeHandler(t *testing.T) {
	name := "spreed-app"
	version := "0.8.3"
	handler := MakeWelcomeHandler(name, version)

	r, _ := http.NewRequest("", "", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Expected response status to be %d, but was %d", http.StatusOK, w.Code)
	}

	if w.Body == nil {
		t.Fatal("No response body was written")
	}

	body := make(map[string]string)
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("Failed to unmarshall json: %v", err)
	}

	if key, expected, actual := name, "Welcome", body[name];  expected != actual {
		t.Errorf("Expected key '%s' to have value '%s', but was '%s'", key, expected, actual)
	}

	if key, expected, actual := "version", version, body["version"];  expected != actual {
		t.Errorf("Expected key '%s' to have value '%s', but was '%s'", key, expected, actual)
	}
}
