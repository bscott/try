package tui

import (
	"testing"
)

func TestDeleteWordBack(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		cursor     int
		wantStr    string
		wantCursor int
	}{
		{
			name:       "delete word at end",
			input:      "hello world",
			cursor:     11,
			wantStr:    "hello ",
			wantCursor: 6,
		},
		{
			name:       "delete word in middle",
			input:      "hello world foo",
			cursor:     11,
			wantStr:    "hello  foo",
			wantCursor: 6,
		},
		{
			name:       "cursor at start",
			input:      "hello world",
			cursor:     0,
			wantStr:    "hello world",
			wantCursor: 0,
		},
		{
			name:       "delete with spaces",
			input:      "hello   world",
			cursor:     13,
			wantStr:    "hello   ",
			wantCursor: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotCursor := deleteWordBack(tt.input, tt.cursor)
			if gotStr != tt.wantStr {
				t.Errorf("deleteWordBack() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("deleteWordBack() cursor = %d, want %d", gotCursor, tt.wantCursor)
			}
		})
	}
}

func TestKillLine(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		cursor int
		want   string
	}{
		{
			name:   "kill from middle",
			input:  "hello world",
			cursor: 6,
			want:   "hello ",
		},
		{
			name:   "kill from start",
			input:  "hello world",
			cursor: 0,
			want:   "",
		},
		{
			name:   "kill from end",
			input:  "hello world",
			cursor: 11,
			want:   "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := killLine(tt.input, tt.cursor)
			if got != tt.want {
				t.Errorf("killLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInsertChar(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		cursor     int
		char       rune
		wantStr    string
		wantCursor int
	}{
		{
			name:       "insert at start",
			input:      "hello",
			cursor:     0,
			char:       'x',
			wantStr:    "xhello",
			wantCursor: 1,
		},
		{
			name:       "insert at end",
			input:      "hello",
			cursor:     5,
			char:       'x',
			wantStr:    "hellox",
			wantCursor: 6,
		},
		{
			name:       "insert in middle",
			input:      "hello",
			cursor:     2,
			char:       'x',
			wantStr:    "hexllo",
			wantCursor: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotCursor := insertChar(tt.input, tt.cursor, tt.char)
			if gotStr != tt.wantStr {
				t.Errorf("insertChar() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("insertChar() cursor = %d, want %d", gotCursor, tt.wantCursor)
			}
		})
	}
}

func TestDeleteChar(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		cursor     int
		wantStr    string
		wantCursor int
	}{
		{
			name:       "delete from middle",
			input:      "hello",
			cursor:     3,
			wantStr:    "helo",
			wantCursor: 2,
		},
		{
			name:       "delete from start - no-op",
			input:      "hello",
			cursor:     0,
			wantStr:    "hello",
			wantCursor: 0,
		},
		{
			name:       "delete from end",
			input:      "hello",
			cursor:     5,
			wantStr:    "hell",
			wantCursor: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotCursor := deleteChar(tt.input, tt.cursor)
			if gotStr != tt.wantStr {
				t.Errorf("deleteChar() str = %q, want %q", gotStr, tt.wantStr)
			}
			if gotCursor != tt.wantCursor {
				t.Errorf("deleteChar() cursor = %d, want %d", gotCursor, tt.wantCursor)
			}
		})
	}
}

func TestMoveCursor(t *testing.T) {
	tests := []struct {
		name   string
		cursor int
		delta  int
		maxLen int
		want   int
	}{
		{
			name:   "move forward",
			cursor: 2,
			delta:  3,
			maxLen: 10,
			want:   5,
		},
		{
			name:   "move backward",
			cursor: 5,
			delta:  -3,
			maxLen: 10,
			want:   2,
		},
		{
			name:   "clamp at start",
			cursor: 2,
			delta:  -5,
			maxLen: 10,
			want:   0,
		},
		{
			name:   "clamp at end",
			cursor: 8,
			delta:  5,
			maxLen: 10,
			want:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := moveCursor(tt.cursor, tt.delta, tt.maxLen)
			if got != tt.want {
				t.Errorf("moveCursor() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid name",
			input:   "my-project",
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantErr: true,
		},
		{
			name:    "contains slash",
			input:   "my/project",
			wantErr: true,
		},
		{
			name:    "contains backslash",
			input:   "my\\project",
			wantErr: true,
		},
		{
			name:    "valid with spaces",
			input:   "my project",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
