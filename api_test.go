package main

import (
	"encoding/json"
	"html/template"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONEndpointsDeclareUTF8Charset(t *testing.T) {
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
			mediaType, parameters, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
			if err != nil {
				t.Fatalf("parse Content-Type %q: %v", response.Header().Get("Content-Type"), err)
			}
			if mediaType != "application/json" {
				t.Errorf("media type = %q, want application/json", mediaType)
			}
			if parameters["charset"] != "utf-8" {
				t.Errorf("charset = %q, want utf-8 (Content-Type: %q)", parameters["charset"], response.Header().Get("Content-Type"))
			}

			var payload map[string]any
			if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
				t.Fatalf("response is not valid JSON: %v", err)
			}
		})
	}
}
