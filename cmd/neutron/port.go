package neutron

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/syncutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
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
		c := common.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		network, _ := cmd.Flags().GetString("network")
		device_id, _ := cmd.Flags().GetString("device-id")
		host, _ := cmd.Flags().GetString("host")
		noHost, _ := cmd.Flags().GetBool("no-host")

		ports, err := c.ListPort(utility.UrlValues(map[string]string{
			"name":            name,
			"network_id":      network,
			"device_id":       device_id,
			"binding:host_id": host,
		}))
		utility.LogError(err, "list ports failed", true)
		if noHost {
			filteredPort := []neutron.Port{}
			for _, port := range ports {
				if port.BindingHostId != "" {
					continue
				}
				filteredPort = append(filteredPort, port)
			}
			common.PrintPorts(filteredPort, long)
		} else {
			common.PrintPorts(ports, long)
		}
	},
}
var portShow = &cobra.Command{
	Use:   "show <port>",
	Short: "Show port",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().NeutronV2()
		port, err := c.FindPort(args[0])
		utility.LogError(err, "show port failed", true)
		common.PrintPort(*port)
	},
}
var portDelete = &cobra.Command{
	Use:   "delete <port> [port ...]",
	Short: "Delete port(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		force, _ := cmd.Flags().GetBool("force")
		c := common.DefaultClient().NeutronV2()
		syncutils.StartTasks(
			syncutils.TaskOption{
				TaskName: "delete ports",
			},
			args,
			func(item string) error {
				port, err := c.FindPort(item)
				if err != nil {
					return fmt.Errorf("show port %s failed: %v", item, err)
				}
				if !force {
					if port.DeviceId != "" {
						console.Warn("port %s is bound to %s", port.Id, port.DeviceId)
						return nil
					}
				}
				console.Info("Reqeust to delete port %s\n", port.Id)
				err = c.DeletePort(port.Id)
				if err != nil {
					utility.LogError(err, fmt.Sprintf("Delete port %s failed", item), false)
				} else {
					console.Info("Delete port %s success", item)
				}
				return nil
			},
		)
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
