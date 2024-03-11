package networking

import (
	"net/url"
	"strings"
	"sync"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/spf13/cobra"
)

var portPrune = &cobra.Command{
	Use:   "prune",
	Short: "Prune ports",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		ports, err := client.NetworkingClient().PortList(query)
		common.LogError(err, "list ports failed", true)
		filterPorts := []networking.Port{}
		for _, port := range ports {
			if port.BindingHostId != "" {
				continue
			}
			if name != "" && !strings.Contains(port.Name, name) {
				continue
			}
			filterPorts = append(filterPorts, port)
		}
		wg := sync.WaitGroup{}
		wg.Add(len(filterPorts))

		for _, port := range filterPorts {
			go func(port networking.Port, wg *sync.WaitGroup) {
				defer wg.Done()
				logging.Info("delete port %s(%s)", port.Id, port.Name)
				err := client.NetworkingClient().PortDelete(port.Id)
				if err != nil {
					logging.Error("delete port %s failed: %v", port.Id, err)
				}
			}(port, &wg)
		}
		wg.Wait()
	},
}

func init() {
	portPrune.Flags().StringP("name", "n", "", "filter by name")
}
