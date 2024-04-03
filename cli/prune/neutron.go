package prune

import (
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var portPrune = &cobra.Command{
	Use:   "port",
	Short: "Prune unbond port(s)",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient()

		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		logging.Info("list ports")
		ports, err := c.NeutronV2().Ports().List(query)
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
			logging.Info("nothing to do")
			return
		}
		c.PrunePorts(filterPorts)
	},
}

func init() {
	portPrune.Flags().StringP("name", "n", "", "filter by name")
}
