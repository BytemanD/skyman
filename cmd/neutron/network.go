package neutron

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Network = &cobra.Command{Use: "network"}

var networkList = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		networks, err := c.Network().List(query)
		if err != nil {
			fmt.Println(err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "AdminStateUp", AutoColor: true},
				{Name: "Subnets", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Network)
					return strings.Join(p.Subnets, "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Shared"}, {Name: "ProviderNetworkType"},
				{Name: "AvailabilityZones"},
			},
			ColumnConfigs: []table.ColumnConfig{
				{Number: 4, Align: text.AlignRight},
			},
		}
		pt.AddItems(networks)
		if long {
			pt.StyleSeparateRows = true
		}
		common.PrintPrettyTable(pt, long)
	},
}
var networkDelete = &cobra.Command{
	Use:   "delete <network> [network ...]",
	Short: "Delete network(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()

		for _, net := range args {
			fmt.Printf("Reqeust to delete network %s\n", net)
			err := c.Network().Delete(net)
			if err != nil {
				console.Error("Delete network %s failed, %s", net, err)
			}
		}
	},
}
var networkShow = &cobra.Command{
	Use:   "show <network>",
	Short: "Show network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()

		network, err := c.Network().Find(args[0])
		utility.LogError(err, "show network failed", true)
		table := common.PrettyItemTable{
			Item: *network,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "ProviderNetworkType"},
				{Name: "ProviderPhysicalNetwork"},
				{Name: "Status"}, {Name: "AdminStateUp"},
				{Name: "Shared"}, {Name: "Subnets"},
				{Name: "Mtu"},
				{Name: "ProjectId"},
				{Name: "AvailabilityZones"},
				{Name: "CreatedAt"},
			},
		}
		common.PrintPrettyItemTable(table)
	},
}
var networkCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()

		// name, _ := cmd.Flags().GetString("name")
		disable, _ := cmd.Flags().GetBool("disable")
		description, _ := cmd.Flags().GetString("description")
		params := map[string]interface{}{
			"name": args[0],
		}
		if disable {
			params["enable"] = false
		}
		if description != "" {
			params["description"] = description
		}
		network, err := c.Network().Create(params)
		utility.LogError(err, "create network failed", true)
		views.PrintNetwork(*network)
	},
}

func init() {
	networkList.Flags().BoolP("long", "l", false, "List additional fields in output")
	networkList.Flags().StringP("name", "n", "", "Search by router name")

	networkCreate.Flags().String("description", "", "Set network description")
	networkCreate.Flags().Bool("disable", false, "Disable router")

	Network.AddCommand(networkList, networkShow, networkDelete, networkCreate)
	// Network.AddCommand(agentCmd)
}
