package prune

import (
	"github.com/spf13/cobra"
)

var PruneCmd = &cobra.Command{Use: "prune", Short: "prune resources"}

func init() {
	PruneCmd.AddCommand(
		portPrune,
		serverPrune,
		volumePrune,
	)
}
