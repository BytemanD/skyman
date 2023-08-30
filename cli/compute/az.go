package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
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
			switch common.CONF.Format {
			case common.DEFAULT, common.FORMAT_TABLE, common.FORMAT_TABLE_LIGHT:
				printAZInfo(azInfo)
			case common.JSON:
				printAzInfoJson(azInfo)
			case common.YAML:
				printAzInfoYaml(azInfo)
			}
		}
	},
}

func init() {
	// flavor list flags
	azList.Flags().Bool("tree", false, "Show tree view.")

	AZ.AddCommand(azList)

}
