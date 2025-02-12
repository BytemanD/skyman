package common

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type Cli[T any] struct {
	Use           string
	Short         string
	ArgNames      []string
	Flags         func(fs *pflag.FlagSet) T
	FlagsRequired []string
	Run           func(args map[string]string, flags T)
}

func NewCommand[T any](command Cli[T]) *cobra.Command {
	useString := func() string {
		useArgs := []string{command.Use}
		for _, name := range command.ArgNames {
			useArgs = append(useArgs, fmt.Sprintf("<%s>", name))
		}
		return strings.Join(useArgs, " ")
	}

	var flags T

	cmd := &cobra.Command{
		Use:   useString(),
		Short: command.Short,
		Args:  cobra.ExactArgs(len(command.ArgNames)),
		Run: func(cmd *cobra.Command, args []string) {
			argsMap := map[string]string{}
			for i, argName := range command.ArgNames {
				argsMap[argName] = args[i]
			}
			command.Run(argsMap, flags)
		},
	}
	// 设置flags
	flags = command.Flags(cmd.Flags())
	for _, flagName := range command.FlagsRequired {
		cmd.MarkFlagRequired(flagName)
	}
	return cmd
}
