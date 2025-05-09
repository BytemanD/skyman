package nova

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
)

var (
	azListFlags flags.AZListFlags
)
var AZ = &cobra.Command{Use: "availability-zone", Aliases: []string{"az"}, Short: "管理可用区"}

var azList = &cobra.Command{
	Use:   "list",
	Short: "List availability zone",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {

		client := common.DefaultClient()
		azInfo, err := client.NovaV2().ListAZ(nil)
		utility.LogError(err, "list availability zones failed", true)

		if *azListFlags.Tree {
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
	azListFlags = flags.AZListFlags{
		Tree: azList.Flags().Bool("tree", false, "Show tree view."),
	}

	AZ.AddCommand(azList)

}
