package networking

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
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
var subnetCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		// name, _ := cmd.Flags().GetString("name")
		noDhcp, _ := cmd.Flags().GetBool("no-dhcp")
		description, _ := cmd.Flags().GetString("description")
		netIdOrName, _ := cmd.Flags().GetString("network")
		cidr, _ := cmd.Flags().GetString("cidr")
		ipVersion, _ := cmd.Flags().GetInt("ip-version")

		network, err := client.NetworkingClient().NetworkFound(netIdOrName)
		common.LogError(err, "create subnet failed", true)

		params := map[string]interface{}{
			"name":       args[0],
			"cidr":       cidr,
			"network_id": network.Id,
			"ip_version": ipVersion,
		}
		if noDhcp {
			params["enable_dhcp"] = false
		}
		if description != "" {
			params["description"] = description
		}
		subnet, err := client.NetworkingClient().SubnetCreate(params)
		common.LogError(err, "create subnet failed", true)
		cli.PrintSubnet(*subnet)
	},
}
var subnetDelete = &cobra.Command{
	Use:   "delete <subnet> [subnet ...]",
	Short: "Delete subnet(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, subnet := range args {
			fmt.Printf("Reqeust to delete subnet %s\n", subnet)
			err := client.NetworkingClient().SubnetDelete(subnet)
			if err != nil {
				logging.Error("Delete subnet %s failed, %s", subnet, err)
			}
		}
	},
}

func init() {
	subnetList.Flags().BoolP("long", "l", false, "List additional fields in output")
	subnetList.Flags().StringP("name", "n", "", "Search by router name")

	subnetCreate.Flags().String("description", "", "Set subnet description")
	subnetCreate.Flags().String("network", "", "Set subnet description")
	subnetCreate.Flags().String("cidr", "", "Subnet range in CIDR notation")
	subnetCreate.Flags().Bool("no-dhcp", false, "Disable DHCP")
	subnetCreate.Flags().Int("ip-version", 4, "IP version (default is 4).")

	subnetCreate.MarkFlagRequired("network")
	subnetCreate.MarkFlagRequired("cidr")

	Subnet.AddCommand(subnetList, subnetCreate, subnetDelete)
}
