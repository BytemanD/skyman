package networking

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/networking"
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
		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Id", "Name", "Status", "AdminStateUp", "Subnets"},
			LongHeaders: []string{
				"Shared", "NetworkType", "AvailabilityZones"},
			SortBy: []table.SortBy{
				{Name: "Name", Mode: table.Asc},
			},
			ColumnConfigs: []table.ColumnConfig{
				{Number: 4, Align: text.AlignRight},
			},
			Slots: map[string]func(item interface{}) interface{}{
				"Subnets": func(item interface{}) interface{} {
					p, _ := item.(networking.Network)
					return strings.Join(p.Subnets, "\n")
				},
			},
		}
		dataListTable.AddItems(networks)
		if long {
			dataListTable.StyleSeparateRows = true
		}
		common.PrintDataListTable(dataListTable, long)

	},
}

func init() {
	networkList.Flags().BoolP("long", "l", false, "List additional fields in output")
	networkList.Flags().StringP("name", "n", "", "Search by router name")

	Network.AddCommand(networkList)
}
