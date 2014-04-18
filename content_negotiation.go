// Copyright 2014 struktur AG. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package httputils

import (
	"net/http"
	"strings"
)

// AcceptsContentType returns true if contentType would be an acceptable
// response to r, otherwise false.
//
// As such, it is only suitable for endpoints which serve a fixed content type.
func AcceptsContentType(r *http.Request, contentType string) bool {
	values, ok := r.Header[http.CanonicalHeaderKey("Accept")]
	if !ok || values[0] == "" {
		// RFC 2616 14.1 specifies that the absence of an Accept header shall
		// be interpreted as */*.
		return true
	}
	targetContentType := parseMimeType(contentType)
	for _, value := range values {
		for _, raw := range strings.Split(value, ",") {
			accepted := parseMimeType(raw)
			if accepted.Matches(targetContentType) {
				return true
			}
		}
	}

	return false
}

// ContainsContentType returns true if the requests content type matches
// contentType, otherwise false.
//
// Note that contentType may be a wildcard.
func ContainsContentType(r *http.Request, contentType string) bool {
	target := parseMimeType(contentType)
	header := r.Header.Get("Content-Type")
	return target.Matches(parseMimeType(header))
}

type mimeType struct {
	Type, SubType string
}

func parseMimeType(raw string) *mimeType {
	// Since we're not really doing content negotiation, we can safely
	// ignore parameters.
	parts := strings.Split(strings.Trim(raw, " "), ";")
	typeParts := strings.Split(strings.Trim(parts[0], " "), "/")
	subType := ""
	if len(typeParts) > 1 {
		subType = typeParts[1]
	} else if len(typeParts) == 1 && typeParts[0] == "*" {
		// Apparently some old/broken implementations will do this.
		subType = "*"
	}

	return &mimeType{
		typeParts[0], subType,
	}
}

func (mime *mimeType) Matches(target *mimeType) bool {
	// Invalid mime types should never be matched.
	if target.Type == "" || target.SubType == "" {
		return false
	}

	if mime.SubType == "*" {
		return target.Type == mime.Type || mime.Type == "*"
	}
	return target.Type == mime.Type && target.SubType == mime.SubType
}

func (mime *mimeType) String() string {
	return mime.Type + "/" + mime.SubType
}
