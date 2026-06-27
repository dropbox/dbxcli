package cmd

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

type jsonHelpOutputForTest struct {
	OK            bool                    `json:"ok"`
	SchemaVersion string                  `json:"schema_version"`
	Command       string                  `json:"command"`
	Input         jsonHelpInput           `json:"input"`
	Results       []jsonHelpResultForTest `json:"results"`
	Warnings      []jsonWarning           `json:"warnings"`
}

type jsonHelpResultForTest struct {
	Status string              `json:"status"`
	Kind   string              `json:"kind"`
	Input  map[string]any      `json:"input"`
	Result jsonCommandManifest `json:"result"`
}

func TestJSONHelpSupportedForms(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantCommand string
		wantPath    string
		wantResults []string
	}{
		{
			name:        "root help output after",
			args:        []string{"--help", "--output=json"},
			wantCommand: "dbxcli",
			wantPath:    "",
			wantResults: []string{"dbxcli", "completion", "completion bash", "completion fish", "completion powershell", "completion zsh", "help", "login", "logout", "ls", "rm", "team", "team add-member", "team info"},
		},
		{
			name:        "root help output before",
			args:        []string{"--output=json", "--help"},
			wantCommand: "dbxcli",
			wantPath:    "",
			wantResults: []string{"dbxcli", "completion", "completion bash", "completion fish", "completion powershell", "completion zsh", "help", "login", "logout", "ls", "rm", "team", "team add-member", "team info"},
		},
		{
			name:        "command help output after",
			args:        []string{"ls", "--help", "--output=json"},
			wantCommand: "ls",
			wantPath:    "ls",
			wantResults: []string{"ls"},
		},
		{
			name:        "command help output before",
			args:        []string{"ls", "--output=json", "--help"},
			wantCommand: "ls",
			wantPath:    "ls",
			wantResults: []string{"ls"},
		},
		{
			name:        "help command output after",
			args:        []string{"help", "ls", "--output=json"},
			wantCommand: "ls",
			wantPath:    "ls",
			wantResults: []string{"ls"},
		},
		{
			name:        "help command output before",
			args:        []string{"--output=json", "help", "ls"},
			wantCommand: "ls",
			wantPath:    "ls",
			wantResults: []string{"ls"},
		},
		{
			name:        "help command describes help",
			args:        []string{"help", "help", "--output=json"},
			wantCommand: "help",
			wantPath:    "help",
			wantResults: []string{"help"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := executeJSONHelpTestRoot(t, tt.args)
			if err != nil {
				t.Fatalf("Execute returned error: %v\nstderr: %s", err, stderr)
			}
			if stderr != "" {
				t.Fatalf("stderr = %q, want empty", stderr)
			}
			got := decodeJSONHelpOutput(t, stdout)
			if !got.OK {
				t.Fatalf("ok = false, want true")
			}
			if got.SchemaVersion != jsonSchemaVersion {
				t.Fatalf("schema_version = %q, want %q", got.SchemaVersion, jsonSchemaVersion)
			}
			if got.Command != tt.wantCommand {
				t.Fatalf("command = %q, want %q", got.Command, tt.wantCommand)
			}
			if got.Input != (jsonHelpInput{Help: true, Path: tt.wantPath}) {
				t.Fatalf("input = %+v, want help path %q", got.Input, tt.wantPath)
			}
			assertJSONHelpResultPaths(t, got, tt.wantResults)
			for _, result := range got.Results {
				if result.Status != jsonHelpStatusDescribed {
					t.Fatalf("status = %q, want %q", result.Status, jsonHelpStatusDescribed)
				}
				if result.Kind != jsonHelpKindCommand {
					t.Fatalf("kind = %q, want %q", result.Kind, jsonHelpKindCommand)
				}
				if len(result.Input) != 0 {
					t.Fatalf("result input = %+v, want empty object", result.Input)
				}
			}
		})
	}
}

func TestJSONHelpCommandSubtree(t *testing.T) {
	stdout, stderr, err := executeJSONHelpTestRoot(t, []string{"team", "--help", "--output=json"})
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstderr: %s", err, stderr)
	}
	got := decodeJSONHelpOutput(t, stdout)
	assertJSONHelpResultPaths(t, got, []string{"team", "team add-member", "team info"})
}

func TestJSONHelpRealRootManifestIncludesPublicCommands(t *testing.T) {
	RootCmd.InitDefaultHelpCmd()

	var got []string
	for _, cmd := range publicCommandSubtree(RootCmd) {
		got = append(got, jsonCommandManifestFor(cmd).Path)
	}
	want := []string{
		"dbxcli",
		"account",
		"completion",
		"completion bash",
		"completion fish",
		"completion powershell",
		"completion zsh",
		"cp",
		"du",
		"get",
		"help",
		"login",
		"logout",
		"ls",
		"mkdir",
		"mv",
		"put",
		"restore",
		"revs",
		"rm",
		"search",
		"share",
		"share list",
		"share list folder",
		"share list link",
		"share-link",
		"share-link create",
		"share-link download",
		"share-link info",
		"share-link list",
		"share-link revoke",
		"share-link update",
		"team",
		"team add-member",
		"team info",
		"team list-groups",
		"team list-members",
		"team remove-member",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("real root manifest paths = %v, want %v", got, want)
	}
}

func TestJSONHelpManifestFields(t *testing.T) {
	stdout, _, err := executeJSONHelpTestRoot(t, []string{"--help", "--output=json"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	got := decodeJSONHelpOutput(t, stdout)

	ls := jsonHelpManifestByPath(t, got, "ls")
	if !ls.SupportsStructuredOutput {
		t.Fatal("ls supports_structured_output = false, want true")
	}
	assertStringSliceEqual(t, "ls auth modes", ls.AuthModes, []string{"personal", "team-access"})
	if ls.DestructiveLevel != destructiveLevelNone {
		t.Fatalf("ls destructive_level = %q, want none", ls.DestructiveLevel)
	}
	assertJSONHelpFlagNames(t, ls.Flags, []string{"as-member", "help", "include-deleted", "output", "verbose"})
	if jsonHelpHasFlag(ls.Flags, "domain") {
		t.Fatal("hidden domain flag should be omitted")
	}
	outputFlag := jsonHelpFlagByName(t, ls.Flags, "output")
	if !outputFlag.Inherited {
		t.Fatal("output flag should be inherited on ls")
	}
	if outputFlag.Type != "string" || outputFlag.Default != "text" {
		t.Fatalf("output flag = %+v, want string text default", outputFlag)
	}

	login := jsonHelpManifestByPath(t, got, "login")
	if login.SupportsStructuredOutput {
		t.Fatal("login supports_structured_output = true, want false")
	}
	if len(login.AuthModes) != 0 {
		t.Fatalf("login auth_modes = %v, want empty", login.AuthModes)
	}

	logout := jsonHelpManifestByPath(t, got, "logout")
	if !logout.SupportsStructuredOutput {
		t.Fatal("logout supports_structured_output = false, want true")
	}
	if len(logout.AuthModes) != 0 {
		t.Fatalf("logout auth_modes = %v, want empty", logout.AuthModes)
	}

	rm := jsonHelpManifestByPath(t, got, "rm")
	if rm.DestructiveLevel != destructiveLevelDelete {
		t.Fatalf("rm destructive_level = %q, want delete", rm.DestructiveLevel)
	}

	teamInfo := jsonHelpManifestByPath(t, got, "team info")
	assertStringSliceEqual(t, "team info auth modes", teamInfo.AuthModes, []string{"team-manage"})

	teamAdd := jsonHelpManifestByPath(t, got, "team add-member")
	if teamAdd.DestructiveLevel != destructiveLevelAdmin {
		t.Fatalf("team add-member destructive_level = %q, want admin", teamAdd.DestructiveLevel)
	}

	helpFromRoot := jsonHelpManifestByPath(t, got, "help")
	if !strings.Contains(helpFromRoot.Use, "[flags]") {
		t.Fatalf("help use from root manifest = %q, want flags in use line", helpFromRoot.Use)
	}
	stdout, _, err = executeJSONHelpTestRoot(t, []string{"help", "help", "--output=json"})
	if err != nil {
		t.Fatalf("Execute help help returned error: %v", err)
	}
	helpFromCommand := jsonHelpManifestByPath(t, decodeJSONHelpOutput(t, stdout), "help")
	if helpFromCommand.Use != helpFromRoot.Use {
		t.Fatalf("help use differs between manifests: root %q, command %q", helpFromRoot.Use, helpFromCommand.Use)
	}
}

func TestJSONHelpIsDeterministic(t *testing.T) {
	first, _, err := executeJSONHelpTestRoot(t, []string{"--help", "--output=json"})
	if err != nil {
		t.Fatalf("first Execute returned error: %v", err)
	}
	second, _, err := executeJSONHelpTestRoot(t, []string{"--output=json", "--help"})
	if err != nil {
		t.Fatalf("second Execute returned error: %v", err)
	}
	if first != second {
		t.Fatalf("JSON help output differs between equivalent invocations\nfirst: %s\nsecond: %s", first, second)
	}

	got := decodeJSONHelpOutput(t, first)
	for _, result := range got.Results {
		var names []string
		for _, flag := range result.Result.Flags {
			names = append(names, flag.Name)
		}
		sorted := append([]string{}, names...)
		sortStrings(sorted)
		if !reflect.DeepEqual(names, sorted) {
			t.Fatalf("flags for %s are not sorted: %v", result.Result.Path, names)
		}
	}
}

func TestJSONHelpDoesNotRequireAuth(t *testing.T) {
	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	stdout, stderr, err := executeJSONHelpTestRoot(t, []string{"help", "ls", "--output=json"})
	if err != nil {
		t.Fatalf("Execute returned error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	got := decodeJSONHelpOutput(t, stdout)
	assertJSONHelpResultPaths(t, got, []string{"ls"})
}

func TestJSONHelpAuthModeInferenceUsesTopLevelTeam(t *testing.T) {
	root := &cobra.Command{Use: "dbxcli"}
	tools := &cobra.Command{Use: "tools"}
	nestedTeam := &cobra.Command{
		Use:  "team",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(tools)
	tools.AddCommand(nestedTeam)

	assertStringSliceEqual(
		t,
		"nested non-top-level team auth modes",
		commandManifestAuthModes(nestedTeam),
		[]string{"personal", "team-access"},
	)
}

func TestJSONHelpAuthModeAnnotationOverridesInference(t *testing.T) {
	root := &cobra.Command{Use: "dbxcli"}
	cmd := &cobra.Command{
		Use:  "custom",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(cmd)
	setCommandAuthModes(cmd, "team-manage")

	assertStringSliceEqual(
		t,
		"explicit auth modes",
		commandManifestAuthModes(cmd),
		[]string{"team-manage"},
	)
}

func TestJSONHelpAuthBypassRequiresInstalledHelpCommand(t *testing.T) {
	root := &cobra.Command{Use: "dbxcli"}
	root.PersistentFlags().String(outputFlag, "text", "")
	if err := root.PersistentFlags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}

	fakeHelp := &cobra.Command{
		Use:  "help",
		RunE: func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(fakeHelp)
	if commandIsJSONHelp(fakeHelp) {
		t.Fatal("non-annotated command named help should not bypass auth")
	}

	jsonRoot := &cobra.Command{Use: "dbxcli"}
	jsonRoot.PersistentFlags().String(outputFlag, "text", "")
	if err := jsonRoot.PersistentFlags().Set(outputFlag, "json"); err != nil {
		t.Fatal(err)
	}
	jsonHelp := newJSONAwareHelpCommand(jsonRoot)
	jsonRoot.AddCommand(jsonHelp)
	if !commandIsJSONHelp(jsonHelp) {
		t.Fatal("installed JSON-aware help command should bypass auth for JSON help")
	}
}

func TestJSONHelpRawArgsDetection(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{
			name: "root help json",
			args: []string{"--help", "--output=json"},
			want: true,
		},
		{
			name: "command help json",
			args: []string{"share", "list", "link", "--help", "--output=json"},
			want: true,
		},
		{
			name: "help command json",
			args: []string{"--output=json", "help", "share", "list", "link"},
			want: true,
		},
		{
			name: "normal command json",
			args: []string{"share", "list", "link", "--output=json"},
			want: false,
		},
		{
			name: "text help",
			args: []string{"share", "list", "link", "--help"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rawArgsRequestJSONHelp(tt.args); got != tt.want {
				t.Fatalf("rawArgsRequestJSONHelp(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestJSONHelpSuppressesCobraDeprecationWarning(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := &cobra.Command{
		Use:           "dbxcli",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().String(outputFlag, "text", "Output format: text, json")
	deprecated := &cobra.Command{
		Use:        "old",
		Short:      "Old command",
		Deprecated: "use new instead",
		RunE:       func(cmd *cobra.Command, args []string) error { return nil },
	}
	root.AddCommand(deprecated)
	installJSONHelp(root)

	args := []string{"old", "--help", "--output=json"}
	restoreDeprecated := func() {}
	if rawArgsRequestJSONHelp(args) {
		restoreDeprecated = temporarilyClearDeprecatedCommands(root)
	}
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("stderr = %q, want no deprecation warning", got)
	}
	got := decodeJSONHelpOutput(t, stdout.String())
	assertJSONHelpResultPaths(t, got, []string{"old"})
	if deprecated.Deprecated != "" {
		t.Fatalf("deprecated restored too early: %q", deprecated.Deprecated)
	}
	restoreDeprecated()
	if deprecated.Deprecated != "use new instead" {
		t.Fatalf("deprecated = %q, want restored original message", deprecated.Deprecated)
	}
}

func TestJSONHelpForUnsupportedStructuredCommand(t *testing.T) {
	stdout, _, err := executeJSONHelpTestRoot(t, []string{"login", "--help", "--output=json"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	got := decodeJSONHelpOutput(t, stdout)
	login := jsonHelpManifestByPath(t, got, "login")
	if login.SupportsStructuredOutput {
		t.Fatal("login supports_structured_output = true, want false")
	}
}

func TestJSONHelpPreservesTextHelpAndRootDefaultHelp(t *testing.T) {
	stdout, stderr, err := executeJSONHelpTestRoot(t, []string{"--help"})
	if err != nil {
		t.Fatalf("text help returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Usage:") {
		t.Fatalf("stdout = %q, want text help", stdout)
	}
	if strings.Contains(stdout, `"ok"`) {
		t.Fatalf("stdout = %q, want no JSON envelope", stdout)
	}

	stdout, stderr, err = executeJSONHelpTestRoot(t, []string{"--output=json"})
	if err != nil {
		t.Fatalf("root default help returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Usage:") || strings.Contains(stdout, `"ok"`) {
		t.Fatalf("stdout = %q, want text root help", stdout)
	}

	stdout, stderr, err = executeJSONHelpTestRoot(t, []string{"help", "help"})
	if err != nil {
		t.Fatalf("help command text help returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "dbxcli help [command]") || strings.Contains(stdout, `"ok"`) {
		t.Fatalf("stdout = %q, want text help-command help", stdout)
	}

	stdout, stderr, err = executeJSONHelpTestRoot(t, []string{"help", "missing"})
	if err != nil {
		t.Fatalf("unknown help topic returned error: %v\nstderr: %s", err, stderr)
	}
	if !strings.Contains(stdout, "Unknown help topic") || !strings.Contains(stdout, "missing") || !strings.Contains(stdout, "Usage:") || strings.Contains(stdout, `"ok"`) {
		t.Fatalf("stdout = %q, want Cobra-style unknown help topic text", stdout)
	}
}

func TestJSONHelpPreservesHelpCommandCompletions(t *testing.T) {
	root := newJSONHelpTestRoot(t)
	root.InitDefaultHelpCmd()
	helpCmd, _, err := root.Find([]string{"help"})
	if err != nil {
		t.Fatalf("find help command: %v", err)
	}
	if helpCmd.ValidArgsFunction == nil {
		t.Fatal("help command should provide command-name completions")
	}

	completions, directive := helpCmd.ValidArgsFunction(helpCmd, nil, "l")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("directive = %v, want no-file-completion", directive)
	}
	want := []cobra.Completion{
		cobra.CompletionWithDesc("login", "Log in and save Dropbox credentials"),
		cobra.CompletionWithDesc("logout", "Log out of the current session"),
		cobra.CompletionWithDesc("ls", "List files and folders"),
	}
	if !reflect.DeepEqual(completions, want) {
		t.Fatalf("completions = %v, want %v", completions, want)
	}

	completions, directive = helpCmd.ValidArgsFunction(helpCmd, nil, "h")
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("help directive = %v, want no-file-completion", directive)
	}
	want = []cobra.Completion{
		cobra.CompletionWithDesc("help", "Help about any command"),
	}
	if !reflect.DeepEqual(completions, want) {
		t.Fatalf("help completions = %v, want %v", completions, want)
	}
}

func TestJSONHelpDoesNotChangeCommandExecution(t *testing.T) {
	var stdout bytes.Buffer
	called := false
	root := newJSONHelpTestRoot(t)
	ls, _, _ := root.Find([]string{"ls"})
	ls.RunE = func(cmd *cobra.Command, args []string) error {
		called = true
		_, _ = cmd.OutOrStdout().Write([]byte("executed\n"))
		return nil
	}
	root.PersistentPreRunE = nil
	root.SetOut(&stdout)
	root.SetArgs([]string{"ls", "--output=json", "/"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatal("ls command did not execute")
	}
	if got := stdout.String(); got != "executed\n" {
		t.Fatalf("stdout = %q, want command output", got)
	}
}

func newJSONHelpTestRoot(t *testing.T) *cobra.Command {
	t.Helper()

	root := &cobra.Command{
		Use:               "dbxcli",
		Short:             "A command line tool for Dropbox users and team admins",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: initDbx,
	}
	root.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logging")
	root.PersistentFlags().String(outputFlag, "text", "Output format: text, json")
	root.PersistentFlags().String("as-member", "", "Member ID to perform action as")
	root.PersistentFlags().String("domain", "", "Override default Dropbox domain, useful for testing")
	if err := root.PersistentFlags().MarkHidden("domain"); err != nil {
		t.Fatalf("hide domain flag: %v", err)
	}

	ls := &cobra.Command{
		Use:   "ls [flags] [<path>]",
		Short: "List files and folders",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	enableStructuredOutput(ls)
	ls.Flags().BoolP("include-deleted", "d", false, "Include deleted files")

	login := &cobra.Command{
		Use:   "login",
		Short: "Log in and save Dropbox credentials",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}

	logout := &cobra.Command{
		Use:   "logout",
		Short: "Log out of the current session",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	enableStructuredOutput(logout)

	rm := &cobra.Command{
		Use:   "rm [flags] <file>",
		Short: "Remove files or folders",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	enableStructuredOutput(rm)
	setCommandDestructiveLevel(rm, destructiveLevelDelete)

	team := &cobra.Command{
		Use:   "team",
		Short: "Team management commands",
	}
	teamInfo := &cobra.Command{
		Use:   "info",
		Short: "Get team information",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	enableStructuredOutput(teamInfo)
	teamAdd := &cobra.Command{
		Use:   "add-member",
		Short: "Add a new member to a team",
		RunE:  func(cmd *cobra.Command, args []string) error { return nil },
	}
	enableStructuredOutput(teamAdd)
	setCommandDestructiveLevel(teamAdd, destructiveLevelAdmin)
	team.AddCommand(teamAdd, teamInfo)

	hidden := &cobra.Command{
		Use:    "hidden",
		Short:  "Hidden command",
		Hidden: true,
		RunE:   func(cmd *cobra.Command, args []string) error { return nil },
	}

	root.AddCommand(ls, login, logout, rm, team, hidden)
	installJSONHelp(root)
	return root
}

func executeJSONHelpTestRoot(t *testing.T, args []string) (string, string, error) {
	t.Helper()

	t.Setenv(envAccessToken, "")
	t.Setenv(envAuthFile, filepath.Join(t.TempDir(), "missing-auth.json"))

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root := newJSONHelpTestRoot(t)
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)

	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func decodeJSONHelpOutput(t *testing.T, stdout string) jsonHelpOutputForTest {
	t.Helper()

	var got jsonHelpOutputForTest
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("decode JSON help output: %v\noutput: %s", err, stdout)
	}
	if got.Warnings == nil {
		t.Fatalf("warnings = nil, want empty array")
	}
	if len(got.Warnings) != 0 {
		t.Fatalf("warnings = %+v, want empty", got.Warnings)
	}
	return got
}

func assertJSONHelpResultPaths(t *testing.T, got jsonHelpOutputForTest, want []string) {
	t.Helper()

	var paths []string
	for _, result := range got.Results {
		paths = append(paths, result.Result.Path)
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("result paths = %v, want %v", paths, want)
	}
}

func jsonHelpManifestByPath(t *testing.T, got jsonHelpOutputForTest, path string) jsonCommandManifest {
	t.Helper()

	for _, result := range got.Results {
		if result.Result.Path == path {
			return result.Result
		}
	}
	t.Fatalf("manifest for %q not found", path)
	return jsonCommandManifest{}
}

func assertJSONHelpFlagNames(t *testing.T, flags []jsonCommandFlag, want []string) {
	t.Helper()

	var got []string
	for _, flag := range flags {
		got = append(got, flag.Name)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("flag names = %v, want %v", got, want)
	}
}

func jsonHelpHasFlag(flags []jsonCommandFlag, name string) bool {
	for _, flag := range flags {
		if flag.Name == name {
			return true
		}
	}
	return false
}

func jsonHelpFlagByName(t *testing.T, flags []jsonCommandFlag, name string) jsonCommandFlag {
	t.Helper()

	for _, flag := range flags {
		if flag.Name == name {
			return flag
		}
	}
	t.Fatalf("flag %q not found", name)
	return jsonCommandFlag{}
}

func sortStrings(values []string) {
	sort.Strings(values)
}
