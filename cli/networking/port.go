package networking

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/networking"
)

var Port = &cobra.Command{Use: "port"}

var portList = &cobra.Command{
	Use:   "list",
	Short: "List ports",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		network, _ := cmd.Flags().GetString("network")
		server, _ := cmd.Flags().GetString("server")
		router, _ := cmd.Flags().GetString("router")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if network != "" {
			query.Set("network_id", network)
		}
		if server != "" {
			query.Set("device_id", server)
		}
		if router != "" {
			query.Set("router_id", router)
		}
		ports := client.Networking.PortList(query)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "MACAddress", Text: "MAC Address"},
				{Name: "FixedIps", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Port)
					ips := []string{}
					if !long {
						for _, fixedIp := range p.FixedIps {
							ips = append(ips, fixedIp.IpAddress)
						}
						return strings.Join(ips, ", ")
					} else {
						for _, fixedIp := range p.FixedIps {
							ips = append(ips,
								fmt.Sprintf("%s@%s", fixedIp.IpAddress, fixedIp.SubnetId))
						}
						return strings.Join(ips, "\n")
					}
				}},
				{Name: "DeviceOwner"},
			},
			LongColumns: []common.Column{
				{Name: "SecurityGroups", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Port)
					return strings.Join(p.SecurityGroups, "\n")
				}},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}
		pt.AddItems(ports)
		common.PrintPrettyTable(pt, long)
		if long {
			pt.StyleSeparateRows = true
		}
		pt.AddItems(ports)
		common.PrintPrettyTable(pt, long)
	},
}
var portShow = &cobra.Command{
	Use:   "show <port>",
	Short: "Show port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		port, err := client.Networking.PortShow(args[0])
		if err != nil {
			common.LogError(err, "show port failed", true)
		}
		table := common.PrettyItemTable{
			Item: *port,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "MACAddress", Text: "MAC Address"},
				{Name: "BindingVnicType"},
				{Name: "BindingVifType"},
				{Name: "BindingDetails", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Port)
					return p.MarshalVifDetails()
				}},
				{Name: "BindingHostId"},
				{Name: "FixedIps"},
				{Name: "DeviceOwner"}, {Name: "DeviceId"},
				{Name: "QosPolicyId"}, {Name: "SecurityGroups"},
				{Name: "RevsionNumber"},
				{Name: "ProjectId"},
				{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			},
		}
		common.PrintPrettyItemTable(table)
	},
}
var portDelete = &cobra.Command{
	Use:   "delete <port> [port ...]",
	Short: "Delete port(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, port := range args {
			fmt.Printf("Reqeust to delete port %s\n", port)
			err := client.Networking.PortDelete(port)
			if err != nil {
				logging.Error("Delete port %s failed, %s", port, err)
			}
		}
	},
}

func init() {
	portList.Flags().BoolP("long", "l", false, "List additional fields in output")
	portList.Flags().StringP("name", "n", "", "Search by router name")
	portList.Flags().String("network", "", "Search by network")
	portList.Flags().String("server", "", "Search by server")

	Port.AddCommand(portList, portShow, portDelete)
}
