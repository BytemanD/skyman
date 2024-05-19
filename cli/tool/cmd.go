package tool

import (
	"github.com/BytemanD/skyman/cli/tool/guest"
	"github.com/BytemanD/skyman/cli/tool/prune"
	"github.com/BytemanD/skyman/cli/tool/server"
	"github.com/spf13/cobra"
)

var ToolCmd = &cobra.Command{Use: "tool", Short: "run specified tool"}

func init() {
	ToolCmd.AddCommand(
		server.ServerCommand,
		guest.GuestCommand,
		server.FlavorCommand,
		prune.PruneCmd,
	)
}
