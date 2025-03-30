package prune

import (
	"net/url"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var portPrune = &cobra.Command{
	Use:   "port",
	Short: "Prune unbond port(s)",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient()

		name, _ := cmd.Flags().GetString("name")
		network, _ := cmd.Flags().GetString("network")
		device, _ := cmd.Flags().GetString("device")
		query := url.Values{}
		if network != "" {
			query.Set("network_id", network)
		}
		if device != "" {
			query.Set("device_id", device)
		}

		console.Info("list ports ...")
		ports, err := c.NeutronV2().ListPort(query)
		utility.LogError(err, "list ports failed", true)
		filterPorts := []neutron.Port{}

		for _, port := range ports {
			if port.BindingVifType != "unbound" || port.DeviceOwner != "" {
				continue
			}
			if name != "" && !strings.Contains(port.Name, name) {
				continue
			}
			filterPorts = append(filterPorts, port)
		}
		if len(filterPorts) == 0 {
			console.Info("all ports is not unbound nothing to do")
			return
		}
		c.PrunePorts(filterPorts)
	},
}

func init() {
	portPrune.Flags().StringP("name", "n", "", "filter by name")
	portPrune.Flags().String("network", "", "filter by network id")
	portPrune.Flags().String("device", "", "filter by device id")
}
