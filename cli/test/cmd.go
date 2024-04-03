package test

import (
	"github.com/spf13/cobra"
)

var TestCmd = &cobra.Command{Use: "test", Short: "Test tools"}

func init() {
	TestCmd.AddCommand(
		attachCmd,
	)
}
