package networking

import (
	"fmt"
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/networking"
)

var Router = &cobra.Command{Use: "router"}

var routerList = &cobra.Command{
	Use:   "list",
	Short: "List routers",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		routers := client.Networking.RouterList(query)
		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Id", "Name", "Status", "AdminStateUp", "Distributed", "HA"},
			LongHeaders:  []string{"Routes", "ExternalGatewayinfo"},
			SortBy: []table.SortBy{
				{Name: "Name", Mode: table.Asc},
			},
			HeaderLabel: map[string]string{"HA": "HA"},
			ColumnConfigs: []table.ColumnConfig{
				{Number: 4, Align: text.AlignRight},
			},
			Slots: map[string]func(item interface{}) interface{}{},
		}

		dataListTable.AddItems(routers)
		common.PrintDataListTable(dataListTable, long)
	},
}
var routerShow = &cobra.Command{
	Use:   "show <router>",
	Short: "Show router",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		router, err := client.Networking.RouterShow(args[0])
		if err != nil {
			common.LogError(err, "show router failed", true)
		}
		table := common.DataTable{
			Item: *router,
			ShortFields: []common.Field{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "AdminStateUp"},
				{Name: "Distributed"}, {Name: "HA", Text: "HA"},
				{Name: "AdminStateUp"},
				{Name: "ExternalGatewayinfo",
					Slot: func(item interface{}) interface{} {
						p, _ := item.(networking.Router)
						return p.MarshalExternalGatewayInfo()
					}},
				{Name: "AvailabilityZones"},
				{Name: "RevsionNumber"},
				{Name: "ProjectId"},
				{Name: "CreatedAt"},
			},
		}
		common.PrintDataTable(table)
	},
}
var routerDelete = &cobra.Command{
	Use:   "delete <router> [router ...]",
	Short: "Delete router(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, router := range args {
			fmt.Printf("Reqeust to delete router %s\n", router)
			err := client.Networking.RouterDelete(router)
			if err != nil {
				logging.Error("Delete router %s failed, %s", router, err)
			}
		}
	},
}

func init() {
	routerList.Flags().BoolP("long", "l", false, "List additional fields in output")
	routerList.Flags().StringP("name", "n", "", "Search by router name")

	Router.AddCommand(routerList, routerShow, routerDelete)
}
