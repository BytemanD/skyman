package compute

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var AZ = &cobra.Command{Use: "az"}

var azList = &cobra.Command{
	Use:   "list",
	Short: "List availability zone",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		tree, _ := cmd.Flags().GetBool("tree")
		client := openstack.DefaultClient()
		azInfo, err := client.NovaV2().AvailabilityZones().Detail(nil)
		utility.LogError(err, "list availability zones failed", true)

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
