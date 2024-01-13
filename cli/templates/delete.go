package templates

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/compute"
)

func deleteFlavor(client *openstack.OpenstackClient, flavor Flavor) {
	computeClient := client.ComputeClient()
	var f *compute.Flavor
	var err error
	if flavor.Id != "" {
		f, err = computeClient.FlavorShow(flavor.Id)
		if err != nil {
			logging.Warning("get flavor %s failed: %s", flavor.Id, err)
			return
		}
	} else if flavor.Name != "" {
		f, err = computeClient.FlavorFoundByName(flavor.Name)
		if err != nil {
			logging.Warning("get flavor %s failed, %s", flavor.Name, err)
			return
		}
	}
	logging.Info("deleting flavor %s", f.Id)
	err = computeClient.FlavorDelete(f.Id)
	common.LogError(err, fmt.Sprintf("delete flavor %s failed", f.Id), false)
}
func deleteNetwork(client *openstack.OpenstackClient, network Network) {
	networkClient := client.NetworkingClient()
	net, err := networkClient.NetworkFound(network.Name)
	if err != nil {
		logging.Warning("get network %s failed: %s", network.Name, err)
		return
	}
	logging.Info("deleting network %s", network.Name)
	err = networkClient.NetworkDelete(net.Id)
	common.LogError(err, fmt.Sprintf("delete network %s failed: %s", network.Name, err), false)
}

func deleteServer(client *openstack.OpenstackClient, server Server, watch bool) {
	computeClient := client.ComputeClient()
	var s *compute.Server
	var err error
	s, err = computeClient.ServerFound(server.Name)
	if s == nil {
		logging.Warning("get server %s failed, %s", server.Name, err)
		return
	}
	logging.Info("deleting server %s", server.Name)
	err = computeClient.ServerDelete(s.Id)
	common.LogError(err, fmt.Sprintf("delete server %s failed", s.Name), false)
	if err == nil && watch {
		client.WaitServerDeleted(s.Id)
	}
}

var DeleteCmd = &cobra.Command{
	Use:   "delete <file>",
	Short: "delete resources of template file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		watch, _ := cmd.Flags().GetBool("watch")
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		common.LogError(err, "load template file failed", true)

		client := cli.GetClient()
		for _, server := range createTemplate.Servers {
			deleteServer(client, server, watch)
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
	DeleteCmd.Flags().Bool("watch", false, "watch the resource progress until it completes.")
}
