package networking

import (
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/spf13/cobra"
)

var Subnet = &cobra.Command{Use: "subnet"}

var subnetList = &cobra.Command{
	Use:   "list",
	Short: "List subnets",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		subnets, err := client.NetworkingClient().SubnetList(query)
		common.LogError(err, "get subnets failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "NetworkId"}, {Name: "Cidr"},
			},
			LongColumns: []common.Column{
				{Name: "EnableDhcp", Text: "Dhcp"},
				{Name: "AllocationPools", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Subnet)
					return strings.Join(p.GetAllocationPoolsList(), ",")
				}},
				{Name: "IpVersion"},
				{Name: "GatewayIp"},
			},
		}
		pt.AddItems(subnets)
		if long {
			pt.StyleSeparateRows = true
		}
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	subnetList.Flags().BoolP("long", "l", false, "List additional fields in output")
	subnetList.Flags().StringP("name", "n", "", "Search by router name")

	Subnet.AddCommand(subnetList)
}
