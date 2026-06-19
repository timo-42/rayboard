package app

import (
	"flag"
	"fmt"
	"io"
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
	return args[0] == "-h" || args[0] == "--help"
}
