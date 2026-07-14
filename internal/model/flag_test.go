package model

import "testing"

func TestMismatchesSlug(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		title    string
		want     bool
	}{
		{"filename matches slug", "retry-semantics-v2.md", "Retry Semantics (v2)", false},
		{"filename differs from slug", "retry.md", "Retry Semantics (v2)", true},
		{"stale filename after retitle", "board-capture.md", "Board Capture Redux", true},
		// a title that reduces to nothing legitimately carries a
		// hand-picked filename, under any name, and is never flagged.
		{"empty-slug title, arbitrary filename", "whatever.md", "ïéK", false},
		{"empty-slug title, another filename", "hand-picked.md", "(!?)", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MismatchesSlug(tt.filename, tt.title); got != tt.want {
				t.Errorf("MismatchesSlug(%q, %q) = %v, want %v", tt.filename, tt.title, got, tt.want)
			}
		})
	}
}
