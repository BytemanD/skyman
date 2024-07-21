package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var AZ = &cobra.Command{Use: "availability-zone", Aliases: []string{"az"}}

var azList = &cobra.Command{
	Use:   "list",
	Short: "List availability zone",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		tree, _ := cmd.Flags().GetBool("tree")
		client := openstack.DefaultClient()
		azInfo, err := client.NovaV2().AZ().Detail(nil)
		utility.LogError(err, "list availability zones failed", true)

		if tree {
			views.PrintAZInfoTree(azInfo)
		} else {
			switch common.CONF.Format {
			case common.DEFAULT, common.FORMAT_TABLE, common.FORMAT_TABLE_LIGHT:
				views.PrintAZInfo(azInfo)
			case common.JSON:
				views.PrintAzInfoJson(azInfo)
			case common.YAML:
				views.PrintAzInfoYaml(azInfo)
			}
		}
	},
}

func init() {
	// flavor list flags
	azList.Flags().Bool("tree", false, "Show tree view.")

	AZ.AddCommand(azList)

}
