package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

type versionOperationOutputForTest struct {
	Input    map[string]any                  `json:"input"`
	Results  []versionOperationResultForTest `json:"results"`
	Warnings []jsonWarning                   `json:"warnings"`
}

type versionOperationResultForTest struct {
	Kind   string         `json:"kind"`
	Input  map[string]any `json:"input"`
	Result versionOutput  `json:"result"`
}

func TestVersionTextUsesCommandOutput(t *testing.T) {
	cmd := NewVersionCommand("1.2.3")
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	stubDropboxVersion(t, "sdk-test", "spec-test")

	if err := versionCommand(cmd, "1.2.3"); err != nil {
		t.Fatalf("versionCommand returned error: %v", err)
	}

	want := "dbxcli version: 1.2.3\nSDK version: sdk-test\nSpec version: spec-test\n"
	if got := stdout.String(); got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestVersionJSONOutputsVersionInfo(t *testing.T) {
	cmd := NewVersionCommand("1.2.3")
	stdout := &bytes.Buffer{}
	cmd.SetOut(stdout)
	cmd.Flags().String(outputFlag, "json", "")
	stubDropboxVersion(t, "sdk-test", "spec-test")

	if err := versionCommand(cmd, "1.2.3"); err != nil {
		t.Fatalf("versionCommand returned error: %v", err)
	}

	var got versionOperationOutputForTest
	if err := json.NewDecoder(bytes.NewReader(stdout.Bytes())).Decode(&got); err != nil {
		t.Fatalf("decode JSON output: %v\noutput: %s", err, stdout.String())
	}
	if got.Input == nil || len(got.Input) != 0 {
		t.Fatalf("input = %#v, want empty object", got.Input)
	}
	if got.Warnings == nil || len(got.Warnings) != 0 {
		t.Fatalf("warnings = %#v, want empty array", got.Warnings)
	}
	if len(got.Results) != 1 {
		t.Fatalf("results length = %d, want 1", len(got.Results))
	}
	result := got.Results[0]
	if result.Kind != versionKindVersion {
		t.Fatalf("kind = %q, want %q", result.Kind, versionKindVersion)
	}
	if result.Input == nil || len(result.Input) != 0 {
		t.Fatalf("result input = %#v, want empty object", result.Input)
	}
	if result.Result.Version != "1.2.3" || result.Result.SDKVersion != "sdk-test" || result.Result.SpecVersion != "spec-test" {
		t.Fatalf("result = %#v, want version info", result.Result)
	}
}

func TestVersionCommandSupportsStructuredOutput(t *testing.T) {
	if !commandSupportsStructuredOutput(NewVersionCommand("1.2.3")) {
		t.Fatal("version command should support structured output")
	}
}

func stubDropboxVersion(t *testing.T, sdkVersion, specVersion string) {
	t.Helper()

	previous := dropboxVersionFunc
	dropboxVersionFunc = func() (string, string) {
		return sdkVersion, specVersion
	}
	t.Cleanup(func() {
		dropboxVersionFunc = previous
	})
}
