package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
)

var Region = &cobra.Command{Use: "region"}

var list = &cobra.Command{
	Use:   "list",
	Short: "List regions",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()
		regions, err := client.Identity.RegionList()
		utility.LogError(err, "list region failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "ParentRegionId"},
				{Name: "Description"},
			},
		}
		pt.AddItems(regions)
		common.PrintPrettyTable(pt, false)
	},
}

func init() {
	Region.AddCommand(list)
}
