package cmd

import "github.com/spf13/cobra"

func parseListOptions(cmd *cobra.Command) (listOptions, error) {
	opts := listOptions{
		long:       optionalBoolFlag(cmd, "long"),
		timeField:  optionalStringFlag(cmd, "time"),
		timeFormat: optionalStringFlag(cmd, "time-format"),
		sortBy:     optionalStringFlag(cmd, "sort"),
		reverse:    optionalBoolFlag(cmd, "reverse"),
		limit:      optionalUint64Flag(cmd, "limit"),
	}
	if err := validateListOptions(opts); err != nil {
		return listOptions{}, err
	}
	return opts, nil
}

func optionalBoolFlag(cmd *cobra.Command, name string) bool {
	if cmd.Flags().Lookup(name) == nil {
		return false
	}
	value, _ := cmd.Flags().GetBool(name)
	return value
}

func optionalStringFlag(cmd *cobra.Command, name string) string {
	if cmd.Flags().Lookup(name) == nil {
		return ""
	}
	value, _ := cmd.Flags().GetString(name)
	return value
}

func optionalUint64Flag(cmd *cobra.Command, name string) uint64 {
	if cmd.Flags().Lookup(name) == nil {
		return 0
	}
	value, _ := cmd.Flags().GetUint64(name)
	return value
}

func validateListOptions(opts listOptions) error {
	if !validListSort(opts.sortBy) {
		return invalidArgumentsErrorfWithDetails("invalid --sort %q (use name, size, time, or type)", flagValueErrorDetails("sort", opts.sortBy), opts.sortBy)
	}
	if !validListTimeField(opts.timeField) {
		return invalidArgumentsErrorfWithDetails("invalid --time %q (use server or client)", flagValueErrorDetails("time", opts.timeField), opts.timeField)
	}
	if !validListTimeFormat(opts.timeFormat) {
		return invalidArgumentsErrorfWithDetails("invalid --time-format %q (use short or rfc3339)", flagValueErrorDetails("time-format", opts.timeFormat), opts.timeFormat)
	}
	return nil
}

func validListSort(sortBy string) bool {
	switch sortBy {
	case "", "name", "size", "time", "type":
		return true
	default:
		return false
	}
}

func validListTimeField(timeField string) bool {
	switch timeField {
	case "", "server", "client":
		return true
	default:
		return false
	}
}

func validListTimeFormat(timeFormat string) bool {
	switch timeFormat {
	case "", "short", "rfc3339":
		return true
	default:
		return false
	}
}
