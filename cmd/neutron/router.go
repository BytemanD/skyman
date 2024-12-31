package neutron

import (
	"net/url"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cmd/views"
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
		router, err := c.Router().Find(args[0])
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
		for _, arg := range args {
			router, err := c.Router().Find(arg)
			if err != nil {
				logging.Warning("get router %s failed", arg)
				continue
			}
			logging.Info("Reqeust to delete router %s\n", arg)
			err = c.Router().Delete(router.Id)
			if err != nil {
				logging.Error("Delete router %s failed, %s", arg, err)
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
		views.PrintRouter(*router)
	},
}

var routerInterface = &cobra.Command{Use: "interface"}

var interfaceAdd = &cobra.Command{
	Use:     "add <router> <interface>",
	Short:   "Add an internal network interface to a router.",
	Example: "  interface add ROUTER <SUBNET>\n  interface add ROUTER port=<PORT>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		r, i := args[0], args[1]

		router, err := c.Router().Find(r)
		utility.LogIfError(err, true, "get router %s failed", r)

		if strings.HasPrefix(i, "port=") {
			i = strings.Replace(i, "port=", "", 1)
			port, err := c.Port().Find(i)
			utility.LogIfError(err, true, "get port,  %s failed", i)
			err = c.Router().AddPort(router.Id, port.Id)
			utility.LogIfError(err, true, "add port,  failed")
			logging.Info("added subnet %s to router %s", i, r)
		} else {
			subnet, err := c.Subnet().Find(i)
			utility.LogIfError(err, true, "get subnet %s failed", i)
			err = c.Router().AddSubnet(router.Id, subnet.Id)
			utility.LogIfError(err, true, "add interface failed")
			logging.Info("added subnet %s to router %s", i, r)

		}
	},
}

var interfaceRemove = &cobra.Command{
	Use:     "remove <router> <interface>",
	Short:   "Remove an internal network interface from a router. interface: <SUBNET>|<port=PORT>",
	Example: "  interface remove ROUTER <SUBNET>\n  interface remove ROUTER port=<PORT>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		r, i := args[0], args[1]

		router, err := c.Router().Find(r)
		utility.LogIfError(err, true, "get router %s failed", r)
		if strings.HasPrefix(i, "port=") {
			i = strings.Replace(i, "port=", "", 1)
			port, err := c.Port().Find(i)
			utility.LogIfError(err, true, "get port %s failed", i)
			err = c.Router().RemovePort(router.Id, port.Id)
			utility.LogIfError(err, true, "remove interface failed")
			logging.Info("remoeved port %s from router %s", i, r)
		} else {
			subnet, err := c.Subnet().Find(i)
			utility.LogIfError(err, true, "get subnet %s failed", i)
			err = c.Router().RemoveSubnet(router.Id, subnet.Id)
			utility.LogIfError(err, true, "remove interface failed")
			logging.Info("removed subnet %s from router %s", i, r)
		}
	},
}
var interfaceList = &cobra.Command{
	Use:   "list <router>",
	Short: "list router interfaces",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		r := args[0]

		router, err := c.Router().Find(r)
		utility.LogIfError(err, true, "get router %s failed", r)

		query := url.Values{}
		query.Set("device_id", router.Id)
		ports, err := c.Port().List(query)
		utility.LogIfError(err, true, "list router ports failed")
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "MACAddress", Text: "MAC Address"},
				{Name: "FixedIps"},
			},
		}
		pt.AddItems(ports)
		common.PrintPrettyTable(pt, false)
	},
}

func init() {
	routerList.Flags().BoolP("long", "l", false, "List additional fields in output")
	routerList.Flags().StringP("name", "n", "", "Search by router name")

	routerCreate.Flags().String("description", "", "Set router description")
	routerCreate.Flags().Bool("disable", false, "Disable router")

	routerInterface.AddCommand(interfaceAdd, interfaceRemove, interfaceList)

	Router.AddCommand(routerList, routerShow, routerDelete, routerCreate,
		routerInterface)
}
