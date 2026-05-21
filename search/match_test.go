package search

import (
	"reflect"
	"testing"
)

func TestIsCaseSensitive(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"", false},
		{"abc", false},
		{"aBc", true},
		{"ABC", true},
		{"123", false},
		{"abc-123", false},
		{"abc-A123", true},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			if got := IsCaseSensitive(tt.query); got != tt.want {
				t.Errorf("IsCaseSensitive(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestSubstringMatch(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		query         string
		caseSensitive bool
		wantMatched   bool
		wantScoreMin  int
	}{
		{
			name:          "simple exact match",
			path:          "/home/user/app.js",
			query:         "app.js",
			caseSensitive: false,
			wantMatched:   true,
			wantScoreMin:  300, // exact filename match bonus
		},
		{
			name:          "case-insensitive match",
			path:          "/home/user/App.js",
			query:         "app.js",
			caseSensitive: false,
			wantMatched:   true,
			wantScoreMin:  300,
		},
		{
			name:          "case-sensitive mismatch",
			path:          "/home/user/App.js",
			query:         "app.js",
			caseSensitive: true,
			wantMatched:   false,
		},
		{
			name:          "boundary bonus match",
			path:          "/home/user/node_modules/express/index.js",
			query:         "express",
			caseSensitive: false,
			wantMatched:   true,
			wantScoreMin:  150, // boundary + base
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, matched := SubstringMatch(tt.path, tt.query, tt.caseSensitive)
			if matched != tt.wantMatched {
				t.Fatalf("SubstringMatch() matched = %v, wantMatched = %v", matched, tt.wantMatched)
			}
			if matched && got.Score < tt.wantScoreMin {
				t.Errorf("SubstringMatch() score = %v, want at least %v", got.Score, tt.wantScoreMin)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		query         string
		caseSensitive bool
		wantMatched   bool
		wantIndices   []int
	}{
		{
			name:          "simple fuzzy match",
			path:          "src/app.js",
			query:         "sajs",
			caseSensitive: false,
			wantMatched:   true,
			wantIndices:   []int{0, 4, 8, 9}, // 's'rc/'a'pp.'j''s' -> indices 0, 4, 8, 9
		},
		{
			name:          "fuzzy mismatch character missing",
			path:          "src/app.js",
			query:         "sax",
			caseSensitive: false,
			wantMatched:   false,
		},
		{
			name:          "case-insensitive fuzzy",
			path:          "src/App.JS",
			query:         "sajs",
			caseSensitive: false,
			wantMatched:   true,
			wantIndices:   []int{0, 4, 8, 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, matched := FuzzyMatch(tt.path, tt.query, tt.caseSensitive)
			if matched != tt.wantMatched {
				t.Fatalf("FuzzyMatch() matched = %v, wantMatched = %v", matched, tt.wantMatched)
			}
			if matched && !reflect.DeepEqual(got.MatchedIndices, tt.wantIndices) {
				t.Errorf("FuzzyMatch() indices = %v, want %v", got.MatchedIndices, tt.wantIndices)
			}
		})
	}
}
