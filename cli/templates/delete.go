package templates

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func deleteFlavor(client *openstack.Openstack, flavor Flavor) {
	var f *nova.Flavor
	var err error
	if flavor.Id != "" {
		f, err = client.NovaV2().Flavors().Show(flavor.Id)
		if err != nil {
			logging.Warning("get flavor %s failed: %s", flavor.Id, err)
			return
		}
	} else if flavor.Name != "" {
		f, err = client.NovaV2().Flavors().Found(flavor.Name)
		if err != nil {
			logging.Warning("get flavor %s failed, %s", flavor.Name, err)
			return
		}
	}
	logging.Info("deleting flavor %s", f.Id)
	err = client.NovaV2().Flavors().Delete(f.Id)
	utility.LogError(err, fmt.Sprintf("delete flavor %s failed", f.Id), false)
}
func deleteNetwork(client *openstack.Openstack, network Network) {
	net, err := client.NeutronV2().Networks().Found(network.Name)
	if err != nil {
		logging.Warning("get network %s failed: %s", network.Name, err)
		return
	}
	logging.Info("deleting network %s", network.Name)
	err = client.NeutronV2().Networks().Delete(net.Id)
	utility.LogError(err, fmt.Sprintf("delete network %s failed: %s", network.Name, err), false)
}

func deleteServer(client *openstack.Openstack, server Server, watch bool) {
	s, err := client.NovaV2().Servers().Found(server.Name)
	if err != nil {
		logging.Warning("get server %s failed, %s", server.Name, err)
		return
	}
	logging.Info("deleting server %s", server.Name)
	err = client.NovaV2().Servers().Delete(s.Id)
	utility.LogError(err, fmt.Sprintf("delete server %s failed", s.Name), false)
	if err == nil && watch {
		client.NovaV2().Servers().WaitDeleted(s.Id)
	}
}

var DeleteCmd = &cobra.Command{
	Use:   "delete <file>",
	Short: "delete resources of template file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wait, _ := cmd.Flags().GetBool("wait")
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		utility.LogError(err, "load template file failed", true)

		client := openstack.DefaultClient()
		for _, server := range createTemplate.Servers {
			deleteServer(client, server, wait)
		}
		for _, network := range createTemplate.Networks {
			deleteNetwork(client, network)
		}
		for _, flavor := range createTemplate.Flavors {
			deleteFlavor(client, flavor)
		}
	},
}

func init() {
	DeleteCmd.Flags().Bool("wait", false, "wait the resource progress until it deleted.")
}
