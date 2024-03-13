package networking

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/openstack/networking"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var portPrune = &cobra.Command{
	Use:   "prune",
	Short: "Prune ports(unbond)",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()
		name, _ := cmd.Flags().GetString("name")
		query := url.Values{}
		logging.Info("list ports")
		ports, err := client.NetworkingClient().PortList(query)
		utility.LogError(err, "list ports failed", true)
		filterPorts := []networking.Port{}
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
		fmt.Printf("即将清理 %d 个Port(s):\n", len(filterPorts))
		for _, port := range filterPorts {
			fmt.Printf("%s (%s)\n", port.Id, port.Name)
		}
		yes := utility.ScanfComfirm("是否清理?", []string{"yes", "y"}, []string{"no", "n"})
		if !yes {
			return
		}
		tg := utility.TaskGroup{
			Func: func(i interface{}) error {
				port := i.(networking.Port)
				logging.Debug("delete port %s(%s)", port.Id, port.Name)
				err := client.NetworkingClient().PortDelete(port.Id)
				if err != nil {
					return fmt.Errorf("delete port %s failed: %v", port.Id, err)
				}
				return nil
			},
			Items:        filterPorts,
			ShowProgress: true,
		}
		err = tg.Start()
		if err != nil {
			logging.Error("清理失败: %v", err)
		} else {
			logging.Info("清理完成")
		}
	},
}

func init() {
	portPrune.Flags().StringP("name", "n", "", "filter by name")
}
