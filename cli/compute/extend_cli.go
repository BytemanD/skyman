package compute

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var hypervisorFlavor = &cobra.Command{Use: "flavor"}

var flavorCapacities = &cobra.Command{
	Use:   "capacities",
	Short: "Get flavor capacities",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		az, _ := cmd.Flags().GetString("az")
		flavor, _ := cmd.Flags().GetString("flavor")

		query := url.Values{}
		if az != "" {
			query.Set("az", "az")
		}
		if flavor != "" {
			query.Set("flavor", flavor)
		}
		capacities, err := client.NovaV2().Hypervisor().FlavorCapacities(query)
		utility.LogError(err, "get flavor capacities failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "FlavorId"}, {Name: "AZ", Text: "Availability Zone"}, {Name: "AllowedSoldNum"},
			},
		}
		pt.AddItems(capacities.Capacities)
		common.PrintPrettyTable(pt, true)
	},
}

func init() {
	flavorCapacities.Flags().String("az", "", "Query by availability zone")
	flavorCapacities.Flags().String("flavor", "", "Query by flavor")
	// TODO: add other options

	hypervisorFlavor.AddCommand(flavorCapacities)
	Hypervisor.AddCommand(hypervisorFlavor)
}
