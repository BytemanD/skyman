package tool

import (
	"github.com/BytemanD/skyman/cmd/tool/guest"
	"github.com/BytemanD/skyman/cmd/tool/neutron"
	"github.com/BytemanD/skyman/cmd/tool/prune"
	"github.com/BytemanD/skyman/cmd/tool/server"
	"github.com/spf13/cobra"
)

var ToolCmd = &cobra.Command{Use: "tool", Short: "run specified tool"}

func init() {
	ToolCmd.AddCommand(
		server.ServerCommand,
		guest.GuestCommand,
		server.FlavorCommand,
		prune.PruneCmd,
		neutron.Vpc,
	)
}
