package output

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestWarnWritesToStderr(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	out := New(&stdout, &stderr, FormatText)
	out.Warn("could not revoke token: %s", "expired")

	if got := stdout.String(); got != "" {
		t.Fatalf("stdout = %q, want empty", got)
	}
	if got, want := stderr.String(), "Warning: could not revoke token: expired\n"; got != want {
		t.Fatalf("stderr = %q, want %q", got, want)
	}
}

func TestRenderText(t *testing.T) {
	var stdout bytes.Buffer
	out := New(&stdout, nil, FormatText)

	err := out.Render(func(w io.Writer) error {
		_, writeErr := w.Write([]byte("plain\n"))
		return writeErr
	}, struct {
		Status string `json:"status"`
	}{Status: "ok"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got, want := stdout.String(), "plain\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRenderTextReturnsRenderError(t *testing.T) {
	wantErr := errors.New("render failed")
	out := New(nil, nil, FormatText)

	err := out.RenderText(func(w io.Writer) error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("RenderText error = %v, want %v", err, wantErr)
	}
}

func TestRenderTextRequiresTextFormat(t *testing.T) {
	out := New(nil, nil, FormatJSON)

	err := out.RenderText(func(w io.Writer) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for text-only renderer in JSON format")
	}
	if !strings.Contains(err.Error(), "requires structured output data") {
		t.Fatalf("error = %q, want structured output data", err.Error())
	}
}

func TestRenderJSON(t *testing.T) {
	var stdout bytes.Buffer
	out := New(&stdout, nil, FormatJSON)

	err := out.Render(func(w io.Writer) error {
		t.Fatal("text renderer should not be called for JSON output")
		return nil
	}, struct {
		Status string `json:"status"`
	}{Status: "ok"})
	if err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if got, want := stdout.String(), "{\"status\":\"ok\"}\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestRenderUnsupportedFormat(t *testing.T) {
	out := New(nil, nil, Format("yaml"))

	err := out.Render(func(w io.Writer) error {
		return nil
	}, nil)
	if err == nil {
		t.Fatal("expected error for unsupported output format")
	}
	if !strings.Contains(err.Error(), "unsupported output format") {
		t.Fatalf("error = %q, want unsupported output format", err.Error())
	}
}
