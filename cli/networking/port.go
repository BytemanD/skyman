package networking

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

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
		client := cli.GetClient()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		network, _ := cmd.Flags().GetString("network")
		device_id, _ := cmd.Flags().GetString("device-id")
		host, _ := cmd.Flags().GetString("host")
		noHost, _ := cmd.Flags().GetBool("no-host")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if network != "" {
			query.Set("network_id", network)
		}
		if device_id != "" {
			query.Set("device_id", device_id)
		}
		if host != "" {
			query.Set("binding:host_id", host)
		}
		ports, err := client.NetworkingClient().PortList(query)
		common.LogError(err, "list ports failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true},
				{Name: "Status", AutoColor: true},
				{Name: "BindingVnicType", Text: "VnicType"},
				{Name: "BindingVifType", Text: "VifType"},
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
				{Name: "SecurityGroups", Slot: func(item interface{}) interface{} {
					p, _ := item.(networking.Port)
					return strings.Join(p.SecurityGroups, "\n")
				}},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}
		if noHost {
			filteredPorts := []networking.Port{}
			for _, port := range ports {
				if port.BindingHostId != "" {
					continue
				}
				filteredPorts = append(filteredPorts, port)
			}
			pt.AddItems(filteredPorts)
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
		client := cli.GetClient()
		port, err := client.NetworkingClient().PortShow(args[0])
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
		wg := sync.WaitGroup{}
		wg.Add(len(args))

		for _, port := range args {
			go func(p string, wg *sync.WaitGroup) {
				defer wg.Done()
				port, err := client.NetworkingClient().PortShow(p)
				if err != nil {
					common.LogError(err, fmt.Sprintf("Show port %s failed", p), false)
					return
				}
				if port.DeviceId != "" {
					logging.Warning("port %s is bound to %s", port.Id, port.DeviceId)
					return
				}
				logging.Info("Reqeust to delete port %s\n", port.Id)
				err = client.NetworkingClient().PortDelete(p)
				if err != nil {
					logging.Error("Delete port %s failed, %s", p, err)
				} else {
					logging.Info("Delete port %s success", p)
				}
			}(port, &wg)
		}
		wg.Wait()
	},
}

func init() {
	portList.Flags().BoolP("long", "l", false, "List additional fields in output")
	portList.Flags().StringP("name", "n", "", "Search by port name")
	portList.Flags().String("network", "", "Search by network")
	portList.Flags().String("device-id", "", "Search by device id")
	portList.Flags().String("host", "", "Search by binding host")
	portList.Flags().Bool("no-host", false, "Search port with no host")

	Port.AddCommand(portList, portShow, portDelete, portPrune)
}
