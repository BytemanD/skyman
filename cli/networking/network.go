package networking

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
)

var Network = &cobra.Command{Use: "network"}

var networkList = &cobra.Command{
	Use:   "list",
	Short: "List networks",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		networks, err := client.Networking.NetworkList(query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "AdminStateUp", AutoColor: true},
				{Name: "Subnets", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Network)
					return strings.Join(p.Subnets, "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Shared"}, {Name: "NetworkType"},
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
		client := cli.GetClient()
		for _, net := range args {
			fmt.Printf("Reqeust to delete network %s\n", net)
			err := client.Networking.NetworkDelete(net)
			if err != nil {
				logging.Error("Delete network %s failed, %s", net, err)
			}
		}
	},
}
var networkShow = &cobra.Command{
	Use:   "show <network>",
	Short: "Show network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		network, err := client.Networking.NetworkShow(args[0])
		if err != nil {
			common.LogError(err, "show network failed", true)
		}
		table := common.PrettyItemTable{
			Item: *network,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "NetworkType"},
				{Name: "PhysicalNetwork"},
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

func init() {
	networkList.Flags().BoolP("long", "l", false, "List additional fields in output")
	networkList.Flags().StringP("name", "n", "", "Search by router name")

	Network.AddCommand(networkList, networkShow, networkDelete)
}
