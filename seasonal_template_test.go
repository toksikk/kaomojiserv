package main

import (
	"bytes"
	"html/template"
	"os"
	"strings"
	"testing"
)

func TestSeasonalTemplateRendersCanonicalKaomoji(t *testing.T) {
	tmpl, err := template.ParseFiles("kaomoji_template.html")
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var rendered bytes.Buffer
	data := templateData{
		Kaomoji:          `<script>alert("nope")</script>`,
		RemainingSeconds: 42,
	}
	if err := tmpl.Execute(&rendered, data); err != nil {
		t.Fatalf("execute template: %v", err)
	}

	html := rendered.String()
	if strings.Contains(html, data.Kaomoji) {
		t.Fatal("kaomoji was rendered without HTML escaping")
	}
	if !strings.Contains(html, `data-kaomoji="&lt;script&gt;alert(&#34;nope&#34;)&lt;/script&gt;"`) {
		t.Fatal("canonical escaped kaomoji is missing from copy data attribute")
	}
	if !strings.Contains(html, `let seconds =  42 ;`) {
		t.Fatal("rotation countdown was not rendered as JavaScript data")
	}
}

func TestSeasonalTemplateIncludesPreviewsAndAccessibility(t *testing.T) {
	contents, err := os.ReadFile("kaomoji_template.html")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}

	markers := []string{
		`'easter', 'april-fools', 'halloween', 'christmas', 'new-year', 'none'`,
		`prefers-reduced-motion: reduce`,
		`aria-live="polite"`,
		`function easterSunday(year)`,
		`function startRave()`,
		`href="/raw"`,
		`href="/api"`,
		`href="/all"`,
		`href="/easter-eggs"`,
		`aria-label="Versteckte Easter-Egg-Anleitung"`,
	}
	for _, marker := range markers {
		if !bytes.Contains(contents, []byte(marker)) {
			t.Errorf("template is missing %q", marker)
		}
	}
}

func TestEasterEggGuideIncludesAllSecrets(t *testing.T) {
	tmpl, err := template.ParseFiles("easter_eggs_template.html")
	if err != nil {
		t.Fatalf("parse Easter egg guide: %v", err)
	}

	var rendered bytes.Buffer
	if err := tmpl.Execute(&rendered, nil); err != nil {
		t.Fatalf("execute Easter egg guide: %v", err)
	}

	guide := rendered.String()
	markers := []string{
		"Three eggs in the margins",
		"bunny",
		"AI-enhancing emotional bandwidth",
		"NICE TRY",
		"YOINK (I SURRENDER)",
		"boo",
		"hohoho",
		"season-new-year-YEAR",
		"↑ ↑ ↓ ↓ ← → ← → B A",
		"exactly seven times",
	}
	for _, marker := range markers {
		if !strings.Contains(guide, marker) {
			t.Errorf("Easter egg guide is missing %q", marker)
		}
	}
}
