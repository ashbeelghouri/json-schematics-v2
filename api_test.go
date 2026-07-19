package schematics

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const apiSchema = `{
  "version": "2.0",
  "global": { "headers": [ {"target":"authorization","required":true} ] },
  "endpoints": [
    { "path":"/users/:id", "method":"POST",
      "query": [ {"target":"verbose","validate":[{"rule":"inOptions","args":{"options":["true","false"]}}]} ],
      "body":  [ {"target":"email","required":true,"validate":[{"rule":"email"}]} ] }
  ]
}`

func TestAPIValidateRequest(t *testing.T) {
	api := NewAPI()
	if err := api.LoadBytes([]byte(apiSchema)); err != nil {
		t.Fatal(err)
	}

	// valid request
	r := httptest.NewRequest("POST", "/users/42?verbose=true", strings.NewReader(`{"email":"a@b.com"}`))
	r.Header.Set("Authorization", "Bearer x")
	if err := api.ValidateRequest(r); err != nil {
		t.Fatalf("expected valid, got %v", err)
	}

	// missing auth header + bad email + bad query
	r2 := httptest.NewRequest("POST", "/users/42?verbose=maybe", strings.NewReader(`{"email":"nope"}`))
	err := api.ValidateRequest(r2)
	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected validation errors, got %v", err)
	}
	if ve.Len() < 3 {
		t.Errorf("expected >=3 errors (auth, query, email), got %d: %v", ve.Len(), ve.Strings("en", "%target:%message"))
	}

	// unmatched route
	r3 := httptest.NewRequest("GET", "/nope", nil)
	if err := api.ValidateRequest(r3); err == nil {
		t.Error("expected route mismatch error")
	}
}

func TestPathMatches(t *testing.T) {
	cases := []struct {
		pat, path string
		want      bool
	}{
		{"/users/:id", "/users/42", true},
		{"/users/:id", "/users/42/extra", false},
		{"/files/*", "/files/a/b/c", true},
		{"/a/b", "/a/b", true},
		{"/a/b", "/a/c", false},
	}
	for _, c := range cases {
		if got := pathMatches(c.pat, c.path); got != c.want {
			t.Errorf("pathMatches(%q,%q)=%v want %v", c.pat, c.path, got, c.want)
		}
	}
	_ = http.MethodGet
}
