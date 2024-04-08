package tool

import (
	"github.com/BytemanD/skyman/cli/tool/guest"
	"github.com/BytemanD/skyman/cli/tool/server"
	"github.com/spf13/cobra"
)

var ToolCmd = &cobra.Command{Use: "tool", Short: "run specified tool"}
var attachCmd = &cobra.Command{Use: "attach", Short: "attach devices to a server"}
var detachCmd = &cobra.Command{Use: "detach", Short: "detach devices from a server"}

func init() {
	ToolCmd.AddCommand(
		server.ServerCommand,
		guest.GuestCommand,
	)
}
