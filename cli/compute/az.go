package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
)

var AZ = &cobra.Command{Use: "az"}

var azList = &cobra.Command{
	Use:   "list",
	Short: "List availability zone",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		tree, _ := cmd.Flags().GetBool("tree")
		client := cli.GetClient()
		azInfo, err := client.Compute.AZListDetail(nil)
		cli.ExitIfError(err)
		if tree {
			printAZInfoTree(azInfo)
		} else {
			printAZInfo(azInfo)
		}
	},
}

func init() {
	// flavor list flags
	azList.Flags().Bool("tree", false, "Show tree view.")

	AZ.AddCommand(azList)

}
