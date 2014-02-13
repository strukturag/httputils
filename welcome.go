// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"encoding/json"
	"net/http"
)

// The conventional path at which the handler returned by MakeWelcomeHandler
// should be mounted.
const WelcomePath = "/welcome"

// MakeWelcomeHandler returns an HTTP handler which renders a JSON response
// containing the provided application name and version in the following format:
//   {
//     "<name>": "Welcome",
//     "version": "<version>"
//   }
func MakeWelcomeHandler(name, version string) http.HandlerFunc {
	response := make(map[string]string)
	response[name] = "Welcome"
	response["version"] = version
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(response)
	}
}
