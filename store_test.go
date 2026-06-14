package notes

import "testing"

func TestExtractSummary(t *testing.T) {
	cases := []struct {
		content string
		want    string
	}{
		{"# Hello\nThis is a note", "This is a note"},
		{"Just plain text", "Just plain text"},
		{"", ""},
		{"# Title\n## Subtitle\nContent here\nMore content", "Content here More content"},
	}
	for _, tc := range cases {
		got := extractSummary(tc.content)
		if got != tc.want {
			t.Errorf("extractSummary(%q) = %q, want %q", tc.content, got, tc.want)
		}
	}
}

func TestNewNoteID_NotEmpty(t *testing.T) {
	id := newNoteID()
	if id == "" {
		t.Error("newNoteID() returned empty string")
	}
	if len(id) != 32 {
		t.Errorf("expected 32 char hex id, got %d chars: %s", len(id), id)
	}
}
