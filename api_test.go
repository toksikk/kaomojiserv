package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"unicode/utf8"
)

func TestJSONEndpointsReturnUTF8(t *testing.T) {
	allk := kaomojis{Kaomojis: []kaomoji{{Kaomoji: "( ˘▽˘)っ♨"}, {Kaomoji: "¯\\_(ツ)_/¯"}}}
	tmpl := template.Must(template.New("main").Parse("ignored"))
	guideTmpl := template.Must(template.New("guide").Parse("ignored"))
	mux := newMux(allk, 60, tmpl, guideTmpl)

	for _, path := range []string{"/api", "/all"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", contentType)
			}
			if !utf8.Valid(response.Body.Bytes()) {
				t.Fatal("response is not valid UTF-8")
			}

			var payload any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}
		})
	}
}
