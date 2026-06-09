package cmd

import "testing"

func TestMvArgValidation(t *testing.T) {
	err := mv(mvCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	err = mv(mvCmd, []string{"/only-one"})
	if err == nil {
		t.Error("expected error for single arg")
	}
}
