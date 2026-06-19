package app

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

func configureLongFlagUsage(flags *flag.FlagSet, output io.Writer, usage string) {
	flags.SetOutput(output)
	flags.Usage = func() {
		fmt.Fprintln(output, usage)
		fmt.Fprintln(output)
		fmt.Fprintln(output, "flags:")
		flags.VisitAll(func(item *flag.Flag) {
			valueName, usageText := flag.UnquoteUsage(item)
			if valueName == "" {
				fmt.Fprintf(output, "  --%s\n", item.Name)
			} else {
				fmt.Fprintf(output, "  --%s %s\n", item.Name, valueName)
			}
			fmt.Fprintf(output, "      %s", usageText)
			if item.DefValue != "" && item.DefValue != "false" {
				fmt.Fprintf(output, " (default %q)", item.DefValue)
			}
			fmt.Fprintln(output)
		})
	}
}

func flagHelpRequested(args []string) bool {
	if len(args) != 1 {
		return false
	}
	return args[0] == "--help"
}

type boolFlag interface {
	IsBoolFlag() bool
}

func rejectSingleDashFlags(flags *flag.FlagSet, args []string) error {
	for index := 0; index < len(args); index++ {
		arg := args[index]
		if arg == "--" {
			return nil
		}
		if arg == "-" || !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			if strings.HasPrefix(arg, "--") {
				name, _, hasInlineValue := strings.Cut(strings.TrimPrefix(arg, "--"), "=")
				item := flags.Lookup(name)
				if item != nil && !hasInlineValue && !isBoolFlag(item) {
					index++
				}
			}
			continue
		}
		if arg == "-h" {
			return fmt.Errorf("invalid flag %q: use --help", arg)
		}
		return fmt.Errorf("invalid flag %q: use --%s", arg, strings.TrimLeft(arg, "-"))
	}
	return nil
}

func isBoolFlag(item *flag.Flag) bool {
	boolValue, ok := item.Value.(boolFlag)
	return ok && boolValue.IsBoolFlag()
}
