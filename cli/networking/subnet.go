package networking

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Subnet = &cobra.Command{Use: "subnet"}

var subnetList = &cobra.Command{
	Use:   "list",
	Short: "List subnets",
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		subnets, err := c.Subnet().List(query)
		utility.LogError(err, "get subnets failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "NetworkId"}, {Name: "Cidr"},
			},
			LongColumns: []common.Column{
				{Name: "EnableDhcp", Text: "Dhcp"},
				{Name: "AllocationPools", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Subnet)
					return strings.Join(p.GetAllocationPoolsList(), ",")
				}},
				{Name: "IpVersion"},
				{Name: "GatewayIp"},
			},
			StyleSeparateRows: long,
		}
		pt.AddItems(subnets)
		common.PrintPrettyTable(pt, long)
	},
}
var subnetCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()

		noDhcp, _ := cmd.Flags().GetBool("no-dhcp")
		description, _ := cmd.Flags().GetString("description")
		netIdOrName, _ := cmd.Flags().GetString("network")
		cidr, _ := cmd.Flags().GetString("cidr")
		ipVersion, _ := cmd.Flags().GetInt("ip-version")

		network, err := c.Network().Found(netIdOrName)
		utility.LogError(err, "get network failed", true)

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
		subnet, err := c.Subnet().Create(params)
		utility.LogError(err, "create subnet failed", true)
		cli.PrintSubnet(*subnet)
	},
}
var subnetShow = &cobra.Command{
	Use:   "show <subnet>",
	Short: "Show subnet",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		subnet, err := c.Subnet().Found(args[0])
		utility.LogError(err, "show subnet failed", true)
		table := common.PrettyItemTable{
			Item: *subnet,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "NetworkId"},
				{Name: "Cidr"},
				{Name: "RevisionNumber"}, {Name: "IpVersion"},
				{Name: "Tags"}, {Name: "EnableDhcp"},
				{Name: "GatewayIp"},
				{Name: "AllocationPools"},
				{Name: "CreatedAt"},
			},
		}
		common.PrintPrettyItemTable(table)

	},
}
var subnetDelete = &cobra.Command{
	Use:   "delete <subnet> [subnet ...]",
	Short: "Delete subnet(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		for _, subnet := range args {
			fmt.Printf("Reqeust to delete subnet %s\n", subnet)
			err := c.Subnet().Delete(subnet)
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

	Subnet.AddCommand(subnetList, subnetCreate, subnetDelete, subnetShow)
}
