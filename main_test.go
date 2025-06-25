package main

import (
	"strings"
	"testing"
)

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name        string
		diff        string
		wantSearch  string
		wantReplace string
		wantErr     bool
		errContains string
	}{
		{
			name: "basic diff parsing",
			diff: `<<<<<<< SEARCH
from flask import Flask
=======
import math
from flask import Flask
>>>>>>> REPLACE`,
			wantSearch:  "from flask import Flask",
			wantReplace: "import math\nfrom flask import Flask",
			wantErr:     false,
		},
		{
			name: "empty replace block",
			diff: `<<<<<<< SEARCH
old code
=======
>>>>>>> REPLACE`,
			wantSearch:  "old code",
			wantReplace: "",
			wantErr:     false,
		},
		{
			name: "multiline search and replace",
			diff: `<<<<<<< SEARCH
def old_function():
    return "old"
=======
def new_function():
    return "new"
    # Added comment
>>>>>>> REPLACE`,
			wantSearch:  "def old_function():\n    return \"old\"",
			wantReplace: "def new_function():\n    return \"new\"\n    # Added comment",
			wantErr:     false,
		},
		{
			name: "diff with extra whitespace",
			diff: `
<<<<<<< SEARCH
test line
=======
replacement line
>>>>>>> REPLACE
`,
			wantSearch:  "test line",
			wantReplace: "replacement line",
			wantErr:     false,
		},
		{
			name: "missing search block",
			diff: `=======
replacement text
>>>>>>> REPLACE`,
			wantSearch:  "",
			wantReplace: "",
			wantErr:     true,
			errContains: "no search block found",
		},
		{
			name:        "missing markers",
			diff:        `just some text without markers`,
			wantSearch:  "",
			wantReplace: "",
			wantErr:     true,
			errContains: "no search block found",
		},
		{
			name: "only search marker",
			diff: `<<<<<<< SEARCH
search text`,
			wantSearch:  "search text",
			wantReplace: "",
			wantErr:     false,
		},
		{
			name: "empty search block",
			diff: `<<<<<<< SEARCH
=======
replacement
>>>>>>> REPLACE`,
			wantSearch:  "",
			wantReplace: "replacement",
			wantErr:     true,
			errContains: "no search block found",
		},
		{
			name: "search with indentation",
			diff: `<<<<<<< SEARCH
    indented code
    more indented
=======
    new indented code
    still indented
>>>>>>> REPLACE`,
			wantSearch:  "    indented code\n    more indented",
			wantReplace: "    new indented code\n    still indented",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSearch, gotReplace, err := parseDiff(tt.diff)

			if tt.wantErr {
				if err == nil {
					t.Errorf("parseDiff() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("parseDiff() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("parseDiff() error = %v, wantErr = false", err)
				return
			}

			if gotSearch != tt.wantSearch {
				t.Errorf("parseDiff() gotSearch = %q, want %q", gotSearch, tt.wantSearch)
			}

			if gotReplace != tt.wantReplace {
				t.Errorf("parseDiff() gotReplace = %q, want %q", gotReplace, tt.wantReplace)
			}
		})
	}
}

func TestPerformEdit(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		searchBlock  string
		replaceBlock string
		want         string
		wantErr      bool
		errContains  string
	}{
		{
			name:         "basic replacement",
			content:      "Hello world\nThis is a test",
			searchBlock:  "Hello world",
			replaceBlock: "Hello universe",
			want:         "Hello universe\nThis is a test",
			wantErr:      false,
		},
		{
			name:         "multiline replacement",
			content:      "line 1\nline 2\nline 3\nline 4",
			searchBlock:  "line 2\nline 3",
			replaceBlock: "new line 2\nnew line 3",
			want:         "line 1\nnew line 2\nnew line 3\nline 4",
			wantErr:      false,
		},
		{
			name:         "deletion (empty replace)",
			content:      "keep this\ndelete this\nkeep this too",
			searchBlock:  "delete this\n",
			replaceBlock: "",
			want:         "keep this\nkeep this too",
			wantErr:      false,
		},
		{
			name:         "insertion (empty search)",
			content:      "line 1\nline 2",
			searchBlock:  "",
			replaceBlock: "inserted line\n",
			want:         "",
			wantErr:      true,
			errContains:  "multiple occurrences",
		},
		{
			name:         "replacement with indentation",
			content:      "def function():\n    old_code()\n    return True",
			searchBlock:  "    old_code()",
			replaceBlock: "    new_code()\n    # Added comment",
			want:         "def function():\n    new_code()\n    # Added comment\n    return True",
			wantErr:      false,
		},
		{
			name:         "search block not found",
			content:      "This is some content",
			searchBlock:  "nonexistent text",
			replaceBlock: "replacement",
			want:         "",
			wantErr:      true,
			errContains:  "search block not found",
		},
		{
			name:         "multiple occurrences",
			content:      "duplicate\nsome text\nduplicate\nmore text",
			searchBlock:  "duplicate",
			replaceBlock: "unique",
			want:         "",
			wantErr:      true,
			errContains:  "multiple occurrences",
		},
		{
			name:         "windows line endings normalization",
			content:      "line 1\r\nline 2\r\nline 3",
			searchBlock:  "line 2\r\n",
			replaceBlock: "new line 2\n",
			want:         "line 1\nnew line 2\nline 3",
			wantErr:      false,
		},
		{
			name:         "mixed line endings",
			content:      "unix\nwindows\r\nmac\rend",
			searchBlock:  "windows\r\n",
			replaceBlock: "normalized\n",
			want:         "unix\nnormalized\nmac\rend",
			wantErr:      false,
		},
		{
			name:         "replace entire content",
			content:      "old content",
			searchBlock:  "old content",
			replaceBlock: "completely new content",
			want:         "completely new content",
			wantErr:      false,
		},
		{
			name:         "empty content with empty search",
			content:      "",
			searchBlock:  "",
			replaceBlock: "new content",
			want:         "",
			wantErr:      true,
			errContains:  "multiple occurrences",
		},
		{
			name:         "empty content with valid search",
			content:      "",
			searchBlock:  "nonexistent",
			replaceBlock: "new content",
			want:         "",
			wantErr:      true,
			errContains:  "search block not found",
		},
		{
			name:         "search at beginning",
			content:      "beginning text\nmiddle\nend",
			searchBlock:  "beginning text",
			replaceBlock: "new beginning",
			want:         "new beginning\nmiddle\nend",
			wantErr:      false,
		},
		{
			name:         "search at end",
			content:      "beginning\nmiddle\nend text",
			searchBlock:  "end text",
			replaceBlock: "new end",
			want:         "beginning\nmiddle\nnew end",
			wantErr:      false,
		},
		{
			name:         "whitespace preservation",
			content:      "  spaced content  \n\ttabbed\n",
			searchBlock:  "  spaced content  ",
			replaceBlock: "  new spaced content  ",
			want:         "  new spaced content  \n\ttabbed\n",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := performEdit(tt.content, tt.searchBlock, tt.replaceBlock)

			if tt.wantErr {
				if err == nil {
					t.Errorf("performEdit() error = nil, wantErr = true")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("performEdit() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("performEdit() error = %v, wantErr = false", err)
				return
			}

			if got != tt.want {
				t.Errorf("performEdit() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPerformEditEdgeCases(t *testing.T) {
	t.Run("very large content", func(t *testing.T) {
		// Test with larger content to ensure performance
		content := strings.Repeat("line of text\n", 1000)
		searchBlock := "line of text\nline of text"
		replaceBlock := "replaced line\nreplaced line"

		_, err := performEdit(content, searchBlock, replaceBlock)
		if err == nil {
			t.Error("expected error due to multiple occurrences, got nil")
		}
		if !strings.Contains(err.Error(), "multiple occurrences") {
			t.Errorf("expected multiple occurrences error, got %v", err)
		}
	})

	t.Run("unicode content", func(t *testing.T) {
		content := "Hello 世界\nこんにちは\n🌍"
		searchBlock := "世界"
		replaceBlock := "world"
		want := "Hello world\nこんにちは\n🌍"

		got, err := performEdit(content, searchBlock, replaceBlock)
		if err != nil {
			t.Errorf("performEdit() error = %v, want nil", err)
		}
		if got != want {
			t.Errorf("performEdit() = %q, want %q", got, want)
		}
	})

	t.Run("special characters", func(t *testing.T) {
		content := "regex.*chars\n[brackets]\n$special"
		searchBlock := "regex.*chars"
		replaceBlock := "normal text"
		want := "normal text\n[brackets]\n$special"

		got, err := performEdit(content, searchBlock, replaceBlock)
		if err != nil {
			t.Errorf("performEdit() error = %v, want nil", err)
		}
		if got != want {
			t.Errorf("performEdit() = %q, want %q", got, want)
		}
	})
}

// Benchmark tests to ensure performance
func BenchmarkParseDiff(b *testing.B) {
	diff := `<<<<<<< SEARCH
from flask import Flask
app = Flask(__name__)
=======
import math
from flask import Flask
app = Flask(__name__)
>>>>>>> REPLACE`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := parseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPerformEdit(b *testing.B) {
	content := strings.Repeat("different line\n", 100)
	searchBlock := "unique search text"
	replaceBlock := "unique replacement text"

	// Create content where search appears only once
	content = "unique start\n" + searchBlock + "\n" + content + "unique end"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := performEdit(content, searchBlock, replaceBlock)
		if err != nil {
			b.Fatal(err)
		}
	}
}
