package httputils

import (
	"net/http"
	"testing"
)

func Test_AcceptsContentType_PassesIfNoHeaderIsPresent(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	if !AcceptsContentType(r, "text/html") {
		t.Error("Expected request without accept header to accept any content")
	}
}

func Test_AcceptsContentType_PassesIfTheHeaderIsBlank(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Accept", "")
	if !AcceptsContentType(r, "text/html") {
		t.Error("Expected request with blank accept header to accept any content")
	}
}

func Test_AcceptContentType_PassesIfTheProvidedContentTypeIsAcceptable(t *testing.T) {
	for _, accept := range []string{
		"text/xml",
		"text/html,text/xml",
		"text/html, text/xml",
		"text/xml;q=0.8",
		"text/xml ;q=0.6",
		"text/*",
		"*/*",
	} {
		r, _ := http.NewRequest("", "", nil)
		r.Header.Add("Accept", accept)
		if !AcceptsContentType(r, "text/xml") {
			t.Errorf("Expected text/xml to match accept header with value %s", accept)
		}
	}
}

func Test_AcceptContentType_FailsIfTheProvidedContentTypeIsNotAcceptable(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Accept", "text/xml")
	if AcceptsContentType(r, "text/html") {
		t.Error("Expected request without any matching content types not to be matched")
	}
}

func Test_AcceptContentType_FailsForInvalidWildcards(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Accept", "*/xml")
	if AcceptsContentType(r, "text/xml") {
		t.Error("Accepted request using invalid wildcard")
	}
}

func Test_AcceptContentType_FailsForInvalidMimeTypes(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Accept", "text")
	if AcceptsContentType(r, "text/xml") {
		t.Error("Invalid match for bad accept header")
	}
}

func Test_AcceptContentType_PassesForInvalidWildcards(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Accept", "*")
	if !AcceptsContentType(r, "text/xml") {
		t.Error("Did not accept invalid wildcard")
	}
}

func Test_ContainsContentType_FailsIfNoContentTypeIsProvided(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	if ContainsContentType(r, "*/*") {
		t.Error("Request without content type was matched")
	}
}

func Test_ContainsContentType_SucceedsForMatchingContentTypes(t *testing.T) {
	r, _ := http.NewRequest("", "", nil)
	r.Header.Add("Content-Type", "text/xml")
	for _, targetContentType := range []string {
		"text/xml",
		"text/*",
		"*/*",
	} {
		if !ContainsContentType(r, targetContentType) {
			t.Errorf("Expected text/xml request to match a content type specification of %s", targetContentType)
		}
	}
}
