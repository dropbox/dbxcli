package cmd

import "testing"

func TestCpArgValidation(t *testing.T) {
	err := cp(cpCmd, []string{})
	if err == nil {
		t.Error("expected error for no args")
	}

	err = cp(cpCmd, []string{"/only-one"})
	if err == nil {
		t.Error("expected error for single arg")
	}
}
