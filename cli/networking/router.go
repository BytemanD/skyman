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
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "AdminStateUp", AutoColor: true}, {Name: "Distributed"},
				{Name: "HA", Text: "HA"},
			},
			LongColumns: []common.Column{
				{Name: "Routes"}, {Name: "ExternalGatewayinfo"},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}
		pt.AddItems(routers)
		common.PrintPrettyTable(pt, long)
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
				{Name: "ExternalGatewayinfo", Slot: func(item interface{}) interface{} {
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
