package cmd

import (
	"github.com/dropbox/dbxcli/v3/internal/jsonschema"
	"github.com/spf13/pflag"
)

const (
	commandManifestVersion                 = "1"
	commandManifestScopeAccuracyBestEffort = "audited_best_effort"

	commandManifestSuccessSchema = "docs/json-schema/v1/success.schema.json"
	commandManifestErrorSchema   = "docs/json-schema/v1/error.schema.json"
	commandManifestContractFile  = "docs/json-schema/v1/commands.json"
	commandManifestCommandSchema = "docs/json-schema/v1/commands.schema.json"
)

type jsonCommandManifestMetadata struct {
	Args           []jsonCommandArg
	Examples       []jsonCommandExample
	Flags          map[string]jsonCommandFlagMetadata
	DropboxScopes  []string
	StdinStdout    jsonCommandStdinStdout
	ResultStatuses []string
	ResultKinds    []string
	WarningCodes   []string
	MayPrompt      bool
	Known          bool
}

type jsonCommandFlagMetadata struct {
	EnumValues []string
	Conflicts  []string
	Required   bool
	Sensitive  bool
	MayPrompt  bool
	ValueKind  string
}

type jsonCommandContractMetadata struct {
	Statuses []string
	Kinds    []string
	Warnings []string
}

var globalCommandFlagMetadata = map[string]jsonCommandFlagMetadata{
	"as-member": {ValueKind: "dropbox_member_id"},
	"help":      {ValueKind: "boolean"},
	"output":    {EnumValues: []string{"text", "json"}, ValueKind: "enum"},
	"verbose":   {ValueKind: "boolean"},
}

var commonListFlagMetadata = map[string]jsonCommandFlagMetadata{
	"limit":       {ValueKind: "integer"},
	"long":        {ValueKind: "boolean"},
	"reverse":     {ValueKind: "boolean"},
	"sort":        {EnumValues: []string{"name", "size", "time", "type"}, ValueKind: "enum"},
	"time":        {EnumValues: []string{"server", "client"}, ValueKind: "enum"},
	"time-format": {EnumValues: []string{"short", "rfc3339"}, ValueKind: "enum"},
}

var revsFlagMetadata = map[string]jsonCommandFlagMetadata{
	"limit":       {ValueKind: "integer"},
	"long":        {ValueKind: "boolean"},
	"time":        {EnumValues: []string{"server", "client"}, ValueKind: "enum"},
	"time-format": {EnumValues: []string{"short", "rfc3339"}, ValueKind: "enum"},
}

var sharedLinkPasswordFlagMetadata = map[string]jsonCommandFlagMetadata{
	"password":        {Conflicts: []string{"password-file", "password-prompt"}, Sensitive: true, ValueKind: "secret"},
	"password-file":   {Conflicts: []string{"password", "password-prompt"}, ValueKind: "local_file"},
	"password-prompt": {Conflicts: []string{"password", "password-file"}, MayPrompt: true, ValueKind: "boolean"},
}

var sharedLinkSettingsFlagMetadata = map[string]jsonCommandFlagMetadata{
	"allow-download":    {Conflicts: []string{"disallow-download"}, ValueKind: "boolean"},
	"audience":          {EnumValues: []string{"public", "team", "members", "no-one"}, ValueKind: "enum"},
	"disallow-download": {Conflicts: []string{"allow-download"}, ValueKind: "boolean"},
	"expires":           {Conflicts: []string{"remove-expiration"}, ValueKind: "rfc3339_timestamp"},
	"remove-expiration": {Conflicts: []string{"expires"}, ValueKind: "boolean"},
}

var commandManifestRegistry = map[string]jsonCommandManifestMetadata{
	"account": {
		Args:          []jsonCommandArg{commandArg("account-id", false, false, "account_id", "Dropbox account ID to look up")},
		Examples:      []jsonCommandExample{{Description: "Display the current account", Command: "dbxcli account"}},
		DropboxScopes: []string{"account_info.read"},
		Known:         true,
	},
	"completion": {
		Examples: []jsonCommandExample{{Description: "Generate Bash completion", Command: "dbxcli completion bash"}},
		Known:    true,
	},
	"completion bash": {
		Flags: map[string]jsonCommandFlagMetadata{"no-descriptions": {ValueKind: "boolean"}},
		Known: true,
	},
	"completion fish": {
		Flags: map[string]jsonCommandFlagMetadata{"no-descriptions": {ValueKind: "boolean"}},
		Known: true,
	},
	"completion powershell": {
		Flags: map[string]jsonCommandFlagMetadata{"no-descriptions": {ValueKind: "boolean"}},
		Known: true,
	},
	"completion zsh": {
		Flags: map[string]jsonCommandFlagMetadata{"no-descriptions": {ValueKind: "boolean"}},
		Known: true,
	},
	"cp": {
		Args: []jsonCommandArg{
			commandArg("source", true, true, "dropbox_path", "One or more Dropbox source paths"),
			commandArg("target", true, false, "dropbox_path", "Dropbox destination path"),
		},
		Examples:      []jsonCommandExample{{Description: "Copy a Dropbox file", Command: "dbxcli cp /from.txt /to.txt"}},
		Flags:         map[string]jsonCommandFlagMetadata{dryRunFlagName: {ValueKind: "boolean"}, "if-exists": {EnumValues: []string{"fail", "skip", "autorename"}, ValueKind: "enum"}},
		DropboxScopes: []string{"files.content.write", "files.metadata.read"},
		Known:         true,
	},
	"du": {
		Examples:      []jsonCommandExample{{Description: "Display space usage", Command: "dbxcli du"}},
		DropboxScopes: []string{"account_info.read"},
		Known:         true,
	},
	"get": {
		Args: []jsonCommandArg{
			commandArg("source", true, false, "dropbox_path", "Dropbox file or folder path"),
			streamCommandArg("target", false, false, "local_path", "Local destination path, or - for stdout"),
		},
		Examples:      []jsonCommandExample{{Description: "Download a file", Command: "dbxcli get /remote.txt ./remote.txt"}},
		Flags:         map[string]jsonCommandFlagMetadata{"recursive": {ValueKind: "boolean"}},
		DropboxScopes: []string{"files.content.read", "files.metadata.read"},
		StdinStdout:   jsonCommandStdinStdout{WritesBinaryStdout: true},
		Known:         true,
	},
	"help": {
		Args:     []jsonCommandArg{commandArg("command", false, true, "command_path", "Command path to describe")},
		Examples: []jsonCommandExample{{Description: "Describe a command as JSON", Command: "dbxcli --output=json help put"}},
		Known:    true,
	},
	"login": {
		Args:      []jsonCommandArg{enumCommandArg("token-type", false, false, "auth_type", "Credential type: personal, team-access, or team-manage", "personal", "team-access", "team-manage")},
		Examples:  []jsonCommandExample{{Description: "Log in with the personal Dropbox app key", Command: "dbxcli login"}},
		Flags:     map[string]jsonCommandFlagMetadata{"app-key": {ValueKind: "dropbox_app_key"}},
		MayPrompt: true,
		Known:     true,
	},
	"logout": {
		Examples: []jsonCommandExample{{Description: "Log out and remove saved credentials", Command: "dbxcli logout"}},
		Known:    true,
	},
	"ls": {
		Args:     []jsonCommandArg{commandArg("path", false, false, "dropbox_path", "Dropbox folder or file path")},
		Examples: []jsonCommandExample{{Description: "List the root folder", Command: "dbxcli ls /"}},
		Flags: mergeCommandFlagMetadata(commonListFlagMetadata, map[string]jsonCommandFlagMetadata{
			"include-deleted": {ValueKind: "boolean"},
			"only-deleted":    {ValueKind: "boolean"},
			"recursive":       {ValueKind: "boolean"},
			"recurse":         {ValueKind: "boolean"},
		}),
		DropboxScopes: []string{"files.metadata.read"},
		Known:         true,
	},
	"mkdir": {
		Args:          []jsonCommandArg{commandArg("directory", true, false, "dropbox_path", "Dropbox directory path to create")},
		Examples:      []jsonCommandExample{{Description: "Create a Dropbox folder", Command: "dbxcli mkdir /Reports"}},
		Flags:         map[string]jsonCommandFlagMetadata{dryRunFlagName: {ValueKind: "boolean"}, "parents": {ValueKind: "boolean"}},
		DropboxScopes: []string{"files.content.write"},
		Known:         true,
	},
	"mv": {
		Args: []jsonCommandArg{
			commandArg("source", true, true, "dropbox_path", "One or more Dropbox source paths"),
			commandArg("target", true, false, "dropbox_path", "Dropbox destination path"),
		},
		Examples:      []jsonCommandExample{{Description: "Move a Dropbox file", Command: "dbxcli mv /from.txt /to.txt"}},
		Flags:         map[string]jsonCommandFlagMetadata{dryRunFlagName: {ValueKind: "boolean"}, "if-exists": {EnumValues: []string{"fail", "skip", "autorename"}, ValueKind: "enum"}},
		DropboxScopes: []string{"files.content.write", "files.metadata.read"},
		Known:         true,
	},
	"put": {
		Args: []jsonCommandArg{
			streamCommandArg("source", true, false, "local_path", "Local source path, or - for stdin"),
			commandArg("target", false, false, "dropbox_path", "Dropbox destination path"),
		},
		Examples: []jsonCommandExample{
			{Description: "Upload a file", Command: "dbxcli put file.txt /destination/file.txt"},
			{Description: "Upload from stdin", Command: "printf 'hello' | dbxcli put - /hello.txt"},
		},
		Flags: map[string]jsonCommandFlagMetadata{
			"chunksize":    {ValueKind: "bytes"},
			"debug":        {ValueKind: "boolean"},
			dryRunFlagName: {ValueKind: "boolean"},
			"if-exists":    {EnumValues: []string{"overwrite", "skip", "fail", "autorename"}, ValueKind: "enum"},
			"recursive":    {ValueKind: "boolean"},
			"workers":      {ValueKind: "integer"},
		},
		DropboxScopes: []string{"files.content.write", "files.metadata.read"},
		StdinStdout:   jsonCommandStdinStdout{ReadsStdin: true},
		Known:         true,
	},
	"restore": {
		Args: []jsonCommandArg{
			commandArg("target-path", true, false, "dropbox_path", "Dropbox file path to restore"),
			commandArg("revision", true, false, "revision", "Dropbox file revision to restore"),
		},
		Examples:      []jsonCommandExample{{Description: "Restore a file revision", Command: "dbxcli restore /Reports/old.pdf 015f"}},
		Flags:         map[string]jsonCommandFlagMetadata{dryRunFlagName: {ValueKind: "boolean"}},
		DropboxScopes: []string{"files.content.write", "files.metadata.read"},
		Known:         true,
	},
	"revs": {
		Args:          []jsonCommandArg{commandArg("file", true, false, "dropbox_path", "Dropbox file path")},
		Examples:      []jsonCommandExample{{Description: "List file revisions", Command: "dbxcli revs /Reports/old.pdf"}},
		Flags:         revsFlagMetadata,
		DropboxScopes: []string{"files.metadata.read"},
		Known:         true,
	},
	"rm": {
		Args:     []jsonCommandArg{commandArg("file", true, true, "dropbox_path", "Dropbox path to remove")},
		Examples: []jsonCommandExample{{Description: "Remove a Dropbox path", Command: "dbxcli rm /old.txt"}},
		Flags: map[string]jsonCommandFlagMetadata{
			dryRunFlagName: {ValueKind: "boolean"},
			"force":        {ValueKind: "boolean"},
			"permanent":    {ValueKind: "boolean"},
			"recursive":    {ValueKind: "boolean"},
		},
		DropboxScopes: []string{"files.content.write", "files.metadata.read"},
		Known:         true,
	},
	"search": {
		Args: []jsonCommandArg{
			commandArg("query", true, false, "string", "Search query"),
			commandArg("path-scope", false, false, "dropbox_path", "Dropbox path scope"),
		},
		Examples: []jsonCommandExample{{Description: "Search by filename", Command: "dbxcli search report /Reports"}},
		Flags: mergeCommandFlagMetadata(commonListFlagMetadata, map[string]jsonCommandFlagMetadata{
			"content":  {ValueKind: "boolean"},
			"order-by": {EnumValues: []string{"relevance", "modified"}, ValueKind: "enum"},
		}),
		DropboxScopes: []string{"files.metadata.read", "files.content.read"},
		Known:         true,
	},
	"share list folder": {
		Examples:      []jsonCommandExample{{Description: "List shared folders", Command: "dbxcli share list folder"}},
		DropboxScopes: []string{"sharing.read"},
		Known:         true,
	},
	"share list link": {
		Args:          []jsonCommandArg{commandArg("path", false, false, "dropbox_path", "Dropbox path filter")},
		Examples:      []jsonCommandExample{{Description: "List shared links", Command: "dbxcli share list link"}},
		DropboxScopes: []string{"sharing.read"},
		Known:         true,
	},
	"share-link create": {
		Args:          []jsonCommandArg{commandArg("path", true, false, "dropbox_path", "Dropbox path to share")},
		Examples:      []jsonCommandExample{{Description: "Create a shared link", Command: "dbxcli share-link create /Reports/report.pdf"}},
		Flags:         mergeCommandFlagMetadata(mergeCommandFlagMetadata(sharedLinkSettingsFlagMetadata, sharedLinkPasswordFlagMetadata), map[string]jsonCommandFlagMetadata{"access": {EnumValues: []string{"viewer", "editor", "max"}, ValueKind: "enum"}}),
		DropboxScopes: []string{"sharing.write", "sharing.read"},
		Known:         true,
	},
	"share-link download": {
		Args: []jsonCommandArg{
			commandArg("url", true, false, "url", "Shared link URL"),
			streamCommandArg("target", false, false, "local_path", "Local destination path, or - for stdout"),
		},
		Examples: []jsonCommandExample{{Description: "Download a shared link", Command: "dbxcli share-link download https://www.dropbox.com/s/example/file.txt"}},
		Flags: mergeCommandFlagMetadata(sharedLinkPasswordFlagMetadata, map[string]jsonCommandFlagMetadata{
			"path":      {ValueKind: "dropbox_path"},
			"recursive": {ValueKind: "boolean"},
		}),
		DropboxScopes: []string{"sharing.read", "files.content.read"},
		StdinStdout:   jsonCommandStdinStdout{WritesBinaryStdout: true},
		Known:         true,
	},
	"share-link info": {
		Args:          []jsonCommandArg{commandArg("url", true, false, "url", "Shared link URL")},
		Examples:      []jsonCommandExample{{Description: "Display shared link metadata", Command: "dbxcli share-link info https://www.dropbox.com/s/example/file.txt"}},
		Flags:         mergeCommandFlagMetadata(sharedLinkPasswordFlagMetadata, map[string]jsonCommandFlagMetadata{"path": {ValueKind: "dropbox_path"}}),
		DropboxScopes: []string{"sharing.read"},
		Known:         true,
	},
	"share-link list": {
		Args:          []jsonCommandArg{commandArg("path", false, false, "dropbox_path", "Dropbox path filter")},
		Examples:      []jsonCommandExample{{Description: "List shared links", Command: "dbxcli share-link list /Reports/report.pdf"}},
		DropboxScopes: []string{"sharing.read"},
		Known:         true,
	},
	"share-link revoke": {
		Args:          []jsonCommandArg{commandArg("url", false, false, "url", "Shared link URL; omit when using --path")},
		Examples:      []jsonCommandExample{{Description: "Revoke a shared link", Command: "dbxcli share-link revoke https://www.dropbox.com/s/example/file.txt"}},
		Flags:         map[string]jsonCommandFlagMetadata{dryRunFlagName: {ValueKind: "boolean"}, "path": {ValueKind: "dropbox_path"}},
		DropboxScopes: []string{"sharing.write", "sharing.read"},
		Known:         true,
	},
	"share-link update": {
		Args:     []jsonCommandArg{commandArg("url", true, false, "url", "Shared link URL")},
		Examples: []jsonCommandExample{{Description: "Set a shared link audience", Command: "dbxcli share-link update https://www.dropbox.com/s/example/file.txt --audience team"}},
		Flags: mergeCommandFlagMetadata(mergeCommandFlagMetadata(sharedLinkSettingsFlagMetadata, sharedLinkPasswordFlagMetadata), map[string]jsonCommandFlagMetadata{
			dryRunFlagName:    {ValueKind: "boolean"},
			"password":        {Conflicts: []string{"password-file", "password-prompt", "remove-password"}, Sensitive: true, ValueKind: "secret"},
			"password-file":   {Conflicts: []string{"password", "password-prompt", "remove-password"}, ValueKind: "local_file"},
			"password-prompt": {Conflicts: []string{"password", "password-file", "remove-password"}, MayPrompt: true, ValueKind: "boolean"},
			"remove-password": {Conflicts: []string{"password", "password-file", "password-prompt"}, ValueKind: "boolean"},
		}),
		DropboxScopes: []string{"sharing.write", "sharing.read"},
		Known:         true,
	},
	"team add-member": {
		Args: []jsonCommandArg{
			commandArg("email", true, false, "email", "Member email address"),
			commandArg("first-name", true, false, "string", "Member first name"),
			commandArg("last-name", true, false, "string", "Member last name"),
		},
		Examples:      []jsonCommandExample{{Description: "Add a team member", Command: "dbxcli team add-member ada@example.com Ada Lovelace"}},
		DropboxScopes: []string{"members.write"},
		Known:         true,
	},
	"team info": {
		Examples:      []jsonCommandExample{{Description: "Display team information", Command: "dbxcli team info"}},
		DropboxScopes: []string{"team_info.read"},
		Known:         true,
	},
	"team list-groups": {
		Examples:      []jsonCommandExample{{Description: "List team groups", Command: "dbxcli team list-groups"}},
		DropboxScopes: []string{"groups.read"},
		Known:         true,
	},
	"team list-members": {
		Examples:      []jsonCommandExample{{Description: "List team members", Command: "dbxcli team list-members"}},
		DropboxScopes: []string{"members.read"},
		Known:         true,
	},
	"team remove-member": {
		Args:          []jsonCommandArg{commandArg("email", true, false, "email", "Member email address")},
		Examples:      []jsonCommandExample{{Description: "Remove a team member", Command: "dbxcli team remove-member ada@example.com"}},
		DropboxScopes: []string{"members.write"},
		Known:         true,
	},
	"version": {
		Examples: []jsonCommandExample{{Description: "Print version information", Command: "dbxcli version"}},
		Known:    true,
	},
}

var commandContractRegistry = map[string]jsonCommandContractMetadata{
	"account":             {Statuses: []string{"found"}, Kinds: []string{"account"}},
	"cp":                  {Statuses: []string{"autorenamed", "copied", "skipped", jsonStatusPlanned}, Kinds: []string{"deleted", "file", "folder"}},
	"du":                  {Statuses: []string{"reported"}, Kinds: []string{"space_usage"}},
	"get":                 {Statuses: []string{"created", "downloaded", "existing"}, Kinds: []string{"file", "folder"}},
	"help":                {Statuses: []string{"described"}, Kinds: []string{"command"}},
	"logout":              {Statuses: []string{"already_logged_out", "logged_out"}, Kinds: []string{"auth"}, Warnings: []string{jsonWarningCodeTokenRevokeFailed}},
	"ls":                  {Statuses: []string{"listed"}, Kinds: []string{"deleted", "file", "folder"}},
	"mkdir":               {Statuses: []string{"created", "existing", jsonStatusPlanned}, Kinds: []string{"folder"}},
	"mv":                  {Statuses: []string{"autorenamed", "moved", "skipped", jsonStatusPlanned}, Kinds: []string{"deleted", "file", "folder"}},
	"put":                 {Statuses: []string{"autorenamed", "created", "existing", "skipped", "uploaded", jsonStatusPlanned}, Kinds: []string{"file", "folder"}, Warnings: []string{jsonWarningCodeSkippedSymlink}},
	"restore":             {Statuses: []string{"restored", jsonStatusPlanned}, Kinds: []string{"file"}},
	"revs":                {Statuses: []string{"revision"}, Kinds: []string{"file"}},
	"rm":                  {Statuses: []string{"deleted", "permanently_deleted", jsonStatusPlanned}, Kinds: []string{"deleted", "file", "folder"}},
	"search":              {Statuses: []string{"found"}, Kinds: []string{"deleted", "file", "folder"}},
	"share list folder":   {Statuses: []string{"listed"}, Kinds: []string{"shared_folder"}},
	"share list link":     {Statuses: []string{"listed"}, Kinds: []string{"file", "folder", "link"}, Warnings: []string{jsonWarningCodeDeprecatedCommand}},
	"share-link create":   {Statuses: []string{"created", "existing"}, Kinds: []string{"file", "folder", "link"}},
	"share-link download": {Statuses: []string{"downloaded"}, Kinds: []string{"file", "folder", "link"}},
	"share-link info":     {Statuses: []string{"found"}, Kinds: []string{"file", "folder", "link"}},
	"share-link list":     {Statuses: []string{"listed"}, Kinds: []string{"file", "folder", "link"}},
	"share-link revoke":   {Statuses: []string{"revoked", jsonStatusPlanned}, Kinds: []string{"file", "folder", "link", "shared_link"}},
	"share-link update":   {Statuses: []string{"updated", jsonStatusPlanned}, Kinds: []string{"file", "folder", "link"}},
	"team add-member":     {Statuses: []string{"added", "completed", "started"}, Kinds: []string{"team_member"}},
	"team info":           {Statuses: []string{"found"}, Kinds: []string{"team"}},
	"team list-groups":    {Statuses: []string{"listed"}, Kinds: []string{"team_group"}},
	"team list-members":   {Statuses: []string{"listed"}, Kinds: []string{"team_member"}},
	"team remove-member":  {Statuses: []string{"completed", "removed", "started"}, Kinds: []string{"team_member"}},
	"version":             {Statuses: []string{"reported"}, Kinds: []string{"version"}},
}

func commandArg(name string, required bool, variadic bool, valueKind string, description string) jsonCommandArg {
	return jsonCommandArg{
		Name:        name,
		Required:    required,
		Variadic:    variadic,
		Placement:   "positional",
		ValueKind:   valueKind,
		Description: description,
		EnumValues:  []string{},
	}
}

func enumCommandArg(name string, required bool, variadic bool, valueKind string, description string, enumValues ...string) jsonCommandArg {
	arg := commandArg(name, required, variadic, valueKind, description)
	arg.EnumValues = sortedCopyStringSlice(enumValues)
	return arg
}

func streamCommandArg(name string, required bool, variadic bool, valueKind string, description string) jsonCommandArg {
	arg := commandArg(name, required, variadic, valueKind, description)
	arg.StreamDash = true
	return arg
}

func commandManifestMetadataFor(path string) jsonCommandManifestMetadata {
	meta := commandManifestRegistry[path]
	meta.Flags = mergeCommandFlagMetadata(globalCommandFlagMetadata, meta.Flags)
	if contract, ok := commandContractRegistry[path]; ok {
		meta.ResultStatuses = sortedCopyStringSlice(contract.Statuses)
		meta.ResultKinds = sortedCopyStringSlice(contract.Kinds)
		meta.WarningCodes = sortedCopyStringSlice(contract.Warnings)
	}
	return meta
}

func mergeCommandFlagMetadata(base, override map[string]jsonCommandFlagMetadata) map[string]jsonCommandFlagMetadata {
	result := make(map[string]jsonCommandFlagMetadata, len(base)+len(override))
	for name, metadata := range base {
		result[name] = metadata
	}
	for name, metadata := range override {
		result[name] = metadata
	}
	return result
}

func commandManifestSchemaRefs(path string, supportsStructuredOutput bool) jsonCommandSchemaRefs {
	refs := jsonCommandSchemaRefs{
		SuccessSchema: commandManifestSuccessSchema,
		ErrorSchema:   commandManifestErrorSchema,
	}
	if supportsStructuredOutput || path == "help" {
		if _, ok := commandContractRegistry[path]; ok {
			refs.CommandContract = commandManifestContractFile + "#/commands/" + path
			refs.CommandSuccessSchema = commandManifestCommandSchema + "#/$defs/" + jsonschema.CommandDefinitionName(path)
		}
	}
	return refs
}

func commandManifestScopeAccuracy(meta jsonCommandManifestMetadata) string {
	if !meta.Known {
		return ""
	}
	return commandManifestScopeAccuracyBestEffort
}

func commandManifestStdinStdout(meta jsonCommandManifestMetadata) jsonCommandStdinStdout {
	stdinStdout := meta.StdinStdout
	if stdinStdout.Stdout == "" {
		stdinStdout.Stdout = "command_results"
	}
	if stdinStdout.WritesBinaryStdout {
		stdinStdout.Stdout = "binary"
	}
	if stdinStdout.Stderr == "" {
		stdinStdout.Stderr = "status_progress_warnings_diagnostics"
	}
	return stdinStdout
}

func commandManifestFlagValueKind(flag *pflag.Flag, metadata jsonCommandFlagMetadata) string {
	if metadata.ValueKind != "" {
		return metadata.ValueKind
	}
	if len(metadata.EnumValues) > 0 {
		return "enum"
	}
	if flag == nil || flag.Value == nil {
		return ""
	}
	switch flag.Value.Type() {
	case "bool":
		return "boolean"
	case "int", "int64", "uint64":
		return "integer"
	default:
		return flag.Value.Type()
	}
}

func normalizeJSONCommandArgs(args []jsonCommandArg) []jsonCommandArg {
	if args == nil {
		return []jsonCommandArg{}
	}
	return args
}

func normalizeJSONCommandExamples(examples []jsonCommandExample) []jsonCommandExample {
	if examples == nil {
		return []jsonCommandExample{}
	}
	return examples
}
