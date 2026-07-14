package model

import "testing"

func TestSlug(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{"spec vector: punctuation and case", "Retry Semantics (v2)", "retry-semantics-v2"},
		// the K below is the Kelvin sign U+212A, not ASCII K: it must be
		// discarded, never case-mapped into ASCII.
		{"spec vector: non-ASCII discarded", "naïve K-scale", "nave-scale"},
		{"plain title", "board capture", "board-capture"},
		{"hyphen runs collapse", "a--b---c", "a-b-c"},
		{"space runs collapse", "a   b", "a-b"},
		{"trim hyphens", " -edges- ", "edges"},
		{"digits kept", "v2 rollout 2026", "v2-rollout-2026"},
		{"empty title", "", ""},
		{"all discarded code points", "ïéK", ""},
		{"only separators", " -- - ", ""},
		{"only punctuation", "(!?)", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slug(tt.title); got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}
