package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dropbox/dbxcli/v3/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	jsonHelpStatusDescribed = "described"
	jsonHelpKindCommand     = "command"

	destructiveLevelNone   = "none"
	destructiveLevelDelete = "delete"
	destructiveLevelAdmin  = "admin"

	commandAuthModesAnnotation        = "dbxcli.authModes"
	commandDestructiveLevelAnnotation = "dbxcli.destructiveLevel"
	commandJSONHelpAnnotation         = "dbxcli.jsonHelpCommand"
)

type jsonHelpInput struct {
	Help bool   `json:"help"`
	Path string `json:"path"`
}

type jsonCommandManifest struct {
	Path                     string                 `json:"path"`
	Use                      string                 `json:"use"`
	Short                    string                 `json:"short"`
	Aliases                  []string               `json:"aliases"`
	Runnable                 bool                   `json:"runnable"`
	Flags                    []jsonCommandFlag      `json:"flags"`
	SupportsStructuredOutput bool                   `json:"supports_structured_output"`
	AuthModes                []string               `json:"auth_modes"`
	DestructiveLevel         string                 `json:"destructive_level"`
	ManifestVersion          string                 `json:"manifest_version"`
	Args                     []jsonCommandArg       `json:"args"`
	Examples                 []jsonCommandExample   `json:"examples"`
	SchemaRefs               jsonCommandSchemaRefs  `json:"schema_refs"`
	DropboxScopes            []string               `json:"dropbox_scopes"`
	ScopeAccuracy            string                 `json:"scope_accuracy"`
	StdinStdout              jsonCommandStdinStdout `json:"stdin_stdout"`
	ResultStatuses           []string               `json:"result_statuses"`
	ResultKinds              []string               `json:"result_kinds"`
	WarningCodes             []string               `json:"warning_codes"`
	MayPrompt                bool                   `json:"may_prompt"`
}

type jsonCommandFlag struct {
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	Default    string   `json:"default"`
	Usage      string   `json:"usage"`
	Inherited  bool     `json:"inherited"`
	Shorthand  string   `json:"shorthand"`
	EnumValues []string `json:"enum_values"`
	Conflicts  []string `json:"conflicts"`
	Required   bool     `json:"required"`
	Sensitive  bool     `json:"sensitive"`
	MayPrompt  bool     `json:"may_prompt"`
	ValueKind  string   `json:"value_kind"`
}

type jsonCommandArg struct {
	Name        string `json:"name"`
	Required    bool   `json:"required"`
	Variadic    bool   `json:"variadic"`
	Placement   string `json:"placement"`
	ValueKind   string `json:"value_kind"`
	Description string `json:"description"`
	StreamDash  bool   `json:"stream_dash"`
}

type jsonCommandExample struct {
	Description string `json:"description"`
	Command     string `json:"command"`
}

type jsonCommandSchemaRefs struct {
	SuccessSchema   string `json:"success_schema"`
	ErrorSchema     string `json:"error_schema"`
	CommandContract string `json:"command_contract,omitempty"`
}

type jsonCommandStdinStdout struct {
	ReadsStdin         bool   `json:"reads_stdin"`
	WritesBinaryStdout bool   `json:"writes_binary_stdout"`
	Stdout             string `json:"stdout"`
	Stderr             string `json:"stderr"`
}

func installJSONHelp(root *cobra.Command) {
	defaultHelp := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		if shouldRenderJSONHelpForHelpFlag(cmd) {
			if err := renderJSONHelp(cmd, cmd); err != nil {
				renderCommandError(cmd, err)
			}
			return
		}
		defaultHelp(cmd, args)
	})
	root.SetHelpCommand(newJSONAwareHelpCommand(root))
}

func newJSONAwareHelpCommand(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long:  "Help provides help for any command in the application.\nSimply type dbxcli help [path to command] for full details.",
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			var completions []cobra.Completion
			target, _, err := root.Find(args)
			if err != nil {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			if target == nil {
				target = root
			}
			for _, child := range target.Commands() {
				if (child.IsAvailableCommand() || child == cmd) && strings.HasPrefix(child.Name(), toComplete) {
					completions = append(completions, cobra.CompletionWithDesc(child.Name(), child.Short))
				}
			}
			return completions, cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _, err := root.Find(args)
			if target == nil || err != nil {
				if shouldRenderJSONHelpForHelpCommand(cmd) {
					if err != nil {
						return err
					}
					return fmt.Errorf("unknown help topic %#q", args)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unknown help topic %#q\n", args)
				return root.Usage()
			}
			if shouldRenderJSONHelpForHelpCommand(cmd) {
				return renderJSONHelp(cmd, target)
			}
			target.InitDefaultHelpFlag()
			target.InitDefaultVersionFlag()
			return target.Help()
		},
	}
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[commandJSONHelpAnnotation] = "true"
	return cmd
}

func shouldRenderJSONHelpForHelpFlag(cmd *cobra.Command) bool {
	if !jsonHelpOutputRequested(cmd) {
		return false
	}
	flag := cmd.Flags().Lookup("help")
	return flag != nil && flag.Changed
}

func shouldRenderJSONHelpForHelpCommand(cmd *cobra.Command) bool {
	return isJSONHelpCommand(cmd) && jsonHelpOutputRequested(cmd)
}

func jsonHelpOutputRequested(cmd *cobra.Command) bool {
	format, err := commandOutputFormatE(cmd)
	return err == nil && format == output.FormatJSON
}

func commandIsJSONHelp(cmd *cobra.Command) bool {
	return shouldRenderJSONHelpForHelpCommand(cmd)
}

func isJSONHelpCommand(cmd *cobra.Command) bool {
	return cmd != nil && cmd.Annotations[commandJSONHelpAnnotation] == "true"
}

func rawArgsRequestJSONHelp(args []string) bool {
	return outputJSONRequested(args) && (rawArgsHaveHelpFlag(args) || rawArgsFirstCommand(args) == "help")
}

func rawArgsHaveHelpFlag(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "--":
			return false
		case "--help", "-h":
			return true
		}
	}
	return false
}

func rawArgsFirstCommand(args []string) string {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--":
			return ""
		case arg == "--output" || arg == "--as-member":
			i++
			continue
		case arg == "--verbose" || arg == "-v" || strings.HasPrefix(arg, "--output=") || strings.HasPrefix(arg, "--as-member="):
			continue
		case strings.HasPrefix(arg, "-"):
			continue
		default:
			return arg
		}
	}
	return ""
}

func temporarilyClearDeprecatedCommands(root *cobra.Command) func() {
	if root == nil {
		return func() {}
	}

	type deprecatedCommand struct {
		cmd        *cobra.Command
		deprecated string
	}

	var changed []deprecatedCommand
	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		if cmd.Deprecated != "" {
			changed = append(changed, deprecatedCommand{
				cmd:        cmd,
				deprecated: cmd.Deprecated,
			})
			cmd.Deprecated = ""
		}
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)

	return func() {
		for _, item := range changed {
			item.cmd.Deprecated = item.deprecated
		}
	}
}

func renderJSONHelp(invocation *cobra.Command, target *cobra.Command) error {
	input := jsonHelpInput{
		Help: true,
		Path: jsonHelpInputPath(target),
	}
	results := jsonHelpOperationResults(target)
	out := newJSONCommandOperationOutput(target, input, results, nil)
	return commandOutput(invocation).Render(nil, out)
}

func jsonHelpInputPath(cmd *cobra.Command) string {
	if cmd == nil || cmd.Root() == cmd {
		return ""
	}
	return jsonCommandPath(cmd)
}

func jsonHelpOperationResults(cmd *cobra.Command) []jsonOperationResult {
	commands := publicCommandSubtree(cmd)
	results := make([]jsonOperationResult, 0, len(commands))
	for _, command := range commands {
		results = append(results, newJSONOperationResult(
			jsonHelpStatusDescribed,
			jsonHelpKindCommand,
			nil,
			jsonCommandManifestFor(command),
		))
	}
	return results
}

func publicCommandSubtree(cmd *cobra.Command) []*cobra.Command {
	if cmd == nil {
		return nil
	}
	var commands []*cobra.Command
	var walk func(*cobra.Command)
	walk = func(current *cobra.Command) {
		if !current.Hidden {
			commands = append(commands, current)
		}
		children := append([]*cobra.Command{}, current.Commands()...)
		sort.Slice(children, func(i, j int) bool {
			return jsonCommandPath(children[i]) < jsonCommandPath(children[j])
		})
		for _, child := range children {
			if child.Hidden {
				continue
			}
			walk(child)
		}
	}
	walk(cmd)
	return commands
}

func jsonCommandManifestFor(cmd *cobra.Command) jsonCommandManifest {
	cmd.InitDefaultHelpFlag()
	cmd.InitDefaultVersionFlag()
	path := jsonManifestCommandPath(cmd)
	meta := commandManifestMetadataFor(path)
	supportsStructuredOutput := commandSupportsStructuredOutput(cmd)

	return jsonCommandManifest{
		Path:                     path,
		Use:                      cmd.UseLine(),
		Short:                    cmd.Short,
		Aliases:                  sortedCopyStringSlice(cmd.Aliases),
		Runnable:                 cmd.Runnable(),
		Flags:                    jsonCommandFlags(cmd, meta.Flags),
		SupportsStructuredOutput: supportsStructuredOutput,
		AuthModes:                commandManifestAuthModes(cmd),
		DestructiveLevel:         commandManifestDestructiveLevel(cmd),
		ManifestVersion:          commandManifestVersion,
		Args:                     normalizeJSONCommandArgs(meta.Args),
		Examples:                 normalizeJSONCommandExamples(meta.Examples),
		SchemaRefs:               commandManifestSchemaRefs(path, supportsStructuredOutput),
		DropboxScopes:            sortedCopyStringSlice(meta.DropboxScopes),
		ScopeAccuracy:            commandManifestScopeAccuracy(meta),
		StdinStdout:              commandManifestStdinStdout(meta),
		ResultStatuses:           sortedCopyStringSlice(meta.ResultStatuses),
		ResultKinds:              sortedCopyStringSlice(meta.ResultKinds),
		WarningCodes:             sortedCopyStringSlice(meta.WarningCodes),
		MayPrompt:                meta.MayPrompt,
	}
}

func jsonManifestCommandPath(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	if cmd.Root() == cmd {
		return cmd.Name()
	}
	return jsonCommandPath(cmd)
}

func jsonCommandFlags(cmd *cobra.Command, metadata map[string]jsonCommandFlagMetadata) []jsonCommandFlag {
	flagsByName := make(map[string]jsonCommandFlag)
	addFlags := func(flags *pflag.FlagSet, inherited bool) {
		if flags == nil {
			return
		}
		flags.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden {
				return
			}
			if _, exists := flagsByName[flag.Name]; exists {
				return
			}
			flagType := ""
			if flag.Value != nil {
				flagType = flag.Value.Type()
			}
			flagMeta := metadata[flag.Name]
			flagsByName[flag.Name] = jsonCommandFlag{
				Name:       flag.Name,
				Type:       flagType,
				Default:    flag.DefValue,
				Usage:      flag.Usage,
				Inherited:  inherited,
				Shorthand:  flag.Shorthand,
				EnumValues: sortedCopyStringSlice(flagMeta.EnumValues),
				Conflicts:  sortedCopyStringSlice(flagMeta.Conflicts),
				Required:   flagMeta.Required,
				Sensitive:  flagMeta.Sensitive,
				MayPrompt:  flagMeta.MayPrompt,
				ValueKind:  commandManifestFlagValueKind(flag, flagMeta),
			}
		})
	}

	addFlags(cmd.NonInheritedFlags(), false)
	addFlags(cmd.InheritedFlags(), true)

	names := make([]string, 0, len(flagsByName))
	for name := range flagsByName {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]jsonCommandFlag, 0, len(names))
	for _, name := range names {
		result = append(result, flagsByName[name])
	}
	return result
}

// CommandManifestFor returns the JSON help manifest for a command.
func CommandManifestFor(cmd *cobra.Command) jsonCommandManifest {
	return jsonCommandManifestFor(cmd)
}

func setCommandAuthModes(cmd *cobra.Command, modes ...string) {
	setCommandAnnotationList(cmd, commandAuthModesAnnotation, modes)
}

func setCommandDestructiveLevel(cmd *cobra.Command, level string) {
	setCommandAnnotationList(cmd, commandDestructiveLevelAnnotation, []string{level})
}

// CommandManifestAuthModes returns the auth modes published in JSON help.
func CommandManifestAuthModes(cmd *cobra.Command) []string {
	return commandManifestAuthModes(cmd)
}

func setCommandAnnotationList(cmd *cobra.Command, key string, values []string) {
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations[key] = strings.Join(values, ",")
}

func commandManifestAuthModes(cmd *cobra.Command) []string {
	if modes := commandAnnotationList(cmd, commandAuthModesAnnotation); modes != nil {
		return modes
	}
	if cmd == nil || cmd.Root() == cmd || !cmd.Runnable() || commandSkipsAuth(cmd) || isNoAuthCommand(cmd) {
		return []string{}
	}
	if commandHasTopLevelName(cmd, "team") {
		return []string{authTokenTypeName(tokenTeamManage)}
	}
	return []string{authTokenTypeName(tokenPersonal), authTokenTypeName(tokenTeamAccess)}
}

func commandManifestDestructiveLevel(cmd *cobra.Command) string {
	if levels := commandAnnotationList(cmd, commandDestructiveLevelAnnotation); len(levels) > 0 {
		return levels[0]
	}
	return destructiveLevelNone
}

func commandAnnotationList(cmd *cobra.Command, key string) []string {
	if cmd == nil || cmd.Annotations == nil {
		return nil
	}
	value, ok := cmd.Annotations[key]
	if !ok {
		return nil
	}
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	sort.Strings(result)
	return result
}

func isNoAuthCommand(cmd *cobra.Command) bool {
	switch cmd.Name() {
	case "login", "logout":
		return true
	default:
		return false
	}
}

func commandHasTopLevelName(cmd *cobra.Command, name string) bool {
	var topLevel *cobra.Command
	for c := cmd; c != nil && c.Parent() != nil; c = c.Parent() {
		topLevel = c
	}
	return topLevel != nil && topLevel.Name() == name
}

func sortedCopyStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	copied := append([]string{}, values...)
	sort.Strings(copied)
	return copied
}
