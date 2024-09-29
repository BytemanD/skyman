package networking

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
)

var Port = &cobra.Command{Use: "port"}

var portList = &cobra.Command{
	Use:   "list",
	Short: "List ports",
	Args: func(cmd *cobra.Command, args []string) (err error) {
		if err = cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		host, _ := cmd.Flags().GetString("host")
		noHost, _ := cmd.Flags().GetBool("no-host")
		if host != "" && noHost {
			return fmt.Errorf("flags --host and --no-host conflict")
		}
		return
	},
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		network, _ := cmd.Flags().GetString("network")
		device_id, _ := cmd.Flags().GetString("device-id")
		host, _ := cmd.Flags().GetString("host")
		noHost, _ := cmd.Flags().GetBool("no-host")

		ports, err := c.Port().List(utility.UrlValues(map[string]string{
			"name":            name,
			"network_id":      network,
			"device_id":       device_id,
			"binding:host_id": host,
		}))
		utility.LogError(err, "list ports failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "BindingVnicType", Text: "VnicType"},
				{Name: "BindingVifType", Text: "VifType"},
				{Name: "MACAddress", Text: "MAC Address"},
				{Name: "FixedIps", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Port)
					ips := []string{}
					if !long {
						for _, fixedIp := range p.FixedIps {
							ips = append(ips, fixedIp.IpAddress)
						}
						return strings.Join(ips, ", ")
					} else {
						data, _ := json.Marshal(p.FixedIps)
						return string(data)
					}
				}},
				{Name: "DeviceOwner"},
				{Name: "BindingHostId"},
			},
			LongColumns: []common.Column{
				{Name: "DeviceId"},
				{Name: "TenantId"},
				{Name: "BindingProfile"},
				{Name: "SecurityGroups", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Port)
					return strings.Join(p.SecurityGroups, "\n")
				}},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}
		if noHost {
			filteredPort := []neutron.Port{}
			for _, port := range ports {
				if port.BindingHostId != "" {
					continue
				}
				filteredPort = append(filteredPort, port)
			}
			pt.AddItems(filteredPort)
		} else {
			pt.AddItems(ports)
		}
		if long {
			pt.StyleSeparateRows = true
		}
		common.PrintPrettyTable(pt, long)
	},
}
var portShow = &cobra.Command{
	Use:   "show <port>",
	Short: "Show port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().NeutronV2()
		port, err := c.Port().Found(args[0])
		if err != nil {
			utility.LogError(err, "show port failed", true)
		}
		table := common.PrettyItemTable{
			Item: *port,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Description"},
				{Name: "Status"},
				{Name: "AdminStateUp"},
				{Name: "MACAddress", Text: "MAC Address"},
				{Name: "BindingVnicType"},
				{Name: "BindingVifType"},
				{Name: "BindingProfile", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Port)
					return p.MarshalBindingProfile()
				}},
				{Name: "BindingDetails", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.Port)
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
		force, _ := cmd.Flags().GetBool("force")
		c := openstack.DefaultClient().NeutronV2()
		tg := syncutils.TaskGroup{
			Func: func(i interface{}) error {
				p := i.(string)
				port, err := c.Port().Show(p)
				if err != nil {
					return fmt.Errorf("show port %s failed: %v", p, err)
				}
				if !force {
					if port.DeviceId != "" {
						logging.Warning("port %s is bound to %s", port.Id, port.DeviceId)
						return nil
					}
				}
				logging.Info("Reqeust to delete port %s\n", port.Id)
				err = c.Port().Delete(p)
				if err != nil {
					utility.LogError(err, fmt.Sprintf("Delete port %s failed", p), false)
				} else {
					logging.Info("Delete port %s success", p)
				}
				return nil
			},
			Items:        args,
			ShowProgress: true,
		}
		tg.Start()
	},
}

func init() {
	portList.Flags().BoolP("long", "l", false, "List additional fields in output")
	portList.Flags().StringP("name", "n", "", "Search by port name")
	portList.Flags().String("network", "", "Search by network")
	portList.Flags().String("device-id", "", "Search by device id")
	portList.Flags().String("host", "", "Search by binding host")
	portList.Flags().Bool("no-host", false, "Search port with no host")

	portDelete.Flags().Bool("force", false, "Force delete")
	Port.AddCommand(portList, portShow, portDelete)
}
