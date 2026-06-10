package cmd

import (
	"fmt"
	"testing"
)

func TestMkdirArgValidation(t *testing.T) {
	err := mkdir(mkdirCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}
}

func TestMkdirTooManyArgs(t *testing.T) {
	err := mkdir(mkdirCmd, []string{"/a", "/b"})
	if err == nil {
		t.Error("expected error for too many args")
	}
}

func TestIsConflictError(t *testing.T) {
	tests := []struct {
		msg  string
		want bool
	}{
		{"path/conflict/folder/", true},
		{"path/conflict/file/", true},
		{"path/not_found/", false},
		{"some other error", false},
	}

	for _, tt := range tests {
		got := isConflictError(fmt.Errorf("%s", tt.msg))
		if got != tt.want {
			t.Errorf("isConflictError(%q) = %v, want %v", tt.msg, got, tt.want)
		}
	}
}
