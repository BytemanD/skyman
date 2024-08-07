package networking

import (
	"fmt"
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
)

var Router = &cobra.Command{Use: "router"}

var routerList = &cobra.Command{
	Use:   "list",
	Short: "List routers",
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		routers, err := c.Router().List(query)
		utility.LogError(err, "list ports failed", true)
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
		c := openstack.DefaultClient().NeutronV2()
		router, err := c.Router().Show(args[0])
		if err != nil {
			utility.LogError(err, "show router failed", true)
		}
		table := common.PrettyItemTable{
			Item: *router,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "AdminStateUp"},
				{Name: "Distributed"}, {Name: "HA", Text: "HA"},
				{Name: "AdminStateUp"},
				{Name: "ExternalGatewayinfo", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Router)
					return p.MarshalExternalGatewayInfo()
				}},
				{Name: "AvailabilityZones"},
				{Name: "RevsionNumber"},
				{Name: "ProjectId"},
				{Name: "CreatedAt"},
			},
		}
		common.PrintPrettyItemTable(table)
	},
}
var routerDelete = &cobra.Command{
	Use:   "delete <router> [router ...]",
	Short: "Delete router(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		for _, router := range args {
			fmt.Printf("Reqeust to delete router %s\n", router)
			err := c.Router().Delete(router)
			if err != nil {
				logging.Error("Delete router %s failed, %s", router, err)
			}
		}
	},
}
var routerCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create router",
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
		router, err := c.Router().Create(params)
		utility.LogError(err, "create router failed", true)
		cli.PrintRouter(*router)
	},
}

func init() {
	routerList.Flags().BoolP("long", "l", false, "List additional fields in output")
	routerList.Flags().StringP("name", "n", "", "Search by router name")

	routerCreate.Flags().String("description", "", "Set router description")
	routerCreate.Flags().Bool("disable", false, "Disable router")

	Router.AddCommand(routerList, routerShow, routerDelete, routerCreate)
}
