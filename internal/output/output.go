package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

type Renderer struct {
	stdout io.Writer
	stderr io.Writer
	format Format
}

func New(stdout, stderr io.Writer, format Format) *Renderer {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	return &Renderer{
		stdout: stdout,
		stderr: stderr,
		format: format,
	}
}

func (r *Renderer) RenderText(render func(io.Writer) error) error {
	if r.format != FormatText {
		return fmt.Errorf("output format %q requires structured output data", r.format)
	}
	return r.Render(render, nil)
}

func (r *Renderer) Render(renderText func(io.Writer) error, data any) error {
	switch r.format {
	case FormatText:
		if renderText == nil {
			return nil
		}
		return renderText(r.stdout)
	case FormatJSON:
		return json.NewEncoder(r.stdout).Encode(data)
	default:
		return fmt.Errorf("unsupported output format %q", r.format)
	}
}

func (r *Renderer) Warn(format string, args ...any) {
	_, _ = fmt.Fprintf(r.stderr, "Warning: "+format+"\n", args...)
}

func (r *Renderer) Error(format string, args ...any) {
	_, _ = fmt.Fprintf(r.stderr, "Error: "+format+"\n", args...)
}

func (r *Renderer) Status(format string, args ...any) {
	_, _ = fmt.Fprintf(r.stderr, format+"\n", args...)
}

func (r *Renderer) Info(format string, args ...any) {
	_, _ = fmt.Fprintf(r.stdout, format+"\n", args...)
}
