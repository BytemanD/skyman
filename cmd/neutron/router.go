package neutron

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
)

var Router = &cobra.Command{Use: "router"}

var routerList = &cobra.Command{
	Use:   "list",
	Short: "List routers",
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		routers, err := c.ListRouter(query)
		utility.LogError(err, "list routers failed", true)
		common.PrintRouters(routers, long)

	},
}
var routerShow = &cobra.Command{
	Use:   "show <router>",
	Short: "Show router",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		router, err := c.FindRouter(args[0])
		utility.LogError(err, "show router failed", true)
		common.PrintRouter(*router)
	},
}
var routerDelete = &cobra.Command{
	Use:   "delete <router> [router ...]",
	Short: "Delete router(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		for _, arg := range args {
			router, err := c.FindRouter(arg)
			if err != nil {
				console.Warn("get router %s failed", arg)
				continue
			}
			console.Info("Reqeust to delete router %s\n", arg)
			err = c.DeleteRouter(router.Id)
			if err != nil {
				console.Error("Delete router %s failed, %s", arg, err)
			}
		}
	},
}
var routerCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create router",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		// name, _ := cmd.Flags().GetString("name")
		disable, _ := cmd.Flags().GetBool("disable")
		description, _ := cmd.Flags().GetString("description")
		params := map[string]any{
			"name": args[0],
		}
		if disable {
			params["enable"] = false
		}
		if description != "" {
			params["description"] = description
		}
		router, err := c.CreateRouter(params)
		utility.LogError(err, "create router failed", true)
		common.PrintRouter(*router)
	},
}

var routerInterface = &cobra.Command{Use: "interface"}

var interfaceAdd = &cobra.Command{
	Use:     "add <router> <interface>",
	Short:   "Add an internal network interface to a router.",
	Example: "  interface add ROUTER <SUBNET>\n  interface add ROUTER port=<PORT>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		r, i := args[0], args[1]

		router, err := c.FindRouter(r)
		utility.LogIfError(err, true, "get router %s failed", r)

		if strings.HasPrefix(i, "port=") {
			i = strings.Replace(i, "port=", "", 1)
			port, err := c.FindPort(i)
			utility.LogIfError(err, true, "get port,  %s failed", i)
			err = c.AddRouterPort(router.Id, port.Id)
			utility.LogIfError(err, true, "add port,  failed")
			console.Info("added subnet %s to router %s", i, r)
		} else {
			subnet, err := c.FindSubnet(i)
			utility.LogIfError(err, true, "get subnet %s failed", i)
			err = c.AddRouterSubnet(router.Id, subnet.Id)
			utility.LogIfError(err, true, "add interface failed")
			console.Info("added subnet %s to router %s", i, r)

		}
	},
}

var interfaceRemove = &cobra.Command{
	Use:     "remove <router> <interface>",
	Short:   "Remove an internal network interface from a router. interface: <SUBNET>|<port=PORT>",
	Example: "  interface remove ROUTER <SUBNET>\n  interface remove ROUTER port=<PORT>",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		r, i := args[0], args[1]

		router, err := c.FindRouter(r)
		utility.LogIfError(err, true, "get router %s failed", r)
		if strings.HasPrefix(i, "port=") {
			i = strings.Replace(i, "port=", "", 1)
			port, err := c.FindPort(i)
			utility.LogIfError(err, true, "get port %s failed", i)
			err = c.RemoveRouterPort(router.Id, port.Id)
			utility.LogIfError(err, true, "remove interface failed")
			console.Info("remoeved port %s from router %s", i, r)
		} else {
			subnet, err := c.FindSubnet(i)
			utility.LogIfError(err, true, "get subnet %s failed", i)
			err = c.RemoveRouterSubnet(router.Id, subnet.Id)
			utility.LogIfError(err, true, "remove interface failed")
			console.Info("removed subnet %s from router %s", i, r)
		}
	},
}
var interfaceList = &cobra.Command{
	Use:   "list <router>",
	Short: "list router interfaces",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		r := args[0]

		router, err := c.FindRouter(r)
		utility.LogIfError(err, true, "get router %s failed", r)

		query := url.Values{}
		query.Set("device_id", router.Id)
		ports, err := c.ListPort(query)
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
