package templates

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func deleteFlavor(client *openstack.Openstack, flavor Flavor) error {
	var f *nova.Flavor
	var err error
	if flavor.Id != "" {
		f, err = client.NovaV2().GetFlavor(flavor.Id)
		if err != nil {
			console.Warn("get flavor %s failed: %s", flavor.Id, err)
			return nil
		}
	} else if flavor.Name != "" {
		f, err = client.NovaV2().FindFlavor(flavor.Name)
		if err != nil {
			console.Warn("get flavor %s failed, %s", flavor.Name, err)
			return nil
		}
	}
	console.Info("deleting flavor %s", f.Id)
	return client.NovaV2().DeleteFlavor(f.Id)
}
func deleteNetwork(client *openstack.Openstack, network Network) error {
	net, err := client.NeutronV2().FindNetwork(network.Name)
	if err != nil {
		console.Warn("get network %s failed: %s", network.Name, err)
		return nil
	}
	console.Info("deleting network %s", network.Name)
	return client.NeutronV2().DeleteNetwork(net.Id)
}

func deleteServer(client *openstack.Openstack, server Server, watch bool) error {
	s, err := client.NovaV2().FindServer(server.Name)
	if err != nil {
		console.Warn("get server %s failed, %s", server.Name, err)
		return nil
	}
	err = client.NovaV2().DeleteServer(s.Id)
	if err != nil {
		return nil
	}

	console.Info("[%s] deleting (%s)", s.Id, s.Name)
	if watch {
		return client.NovaV2().WaitServerDeleted(s.Id)
	}
	return nil
}

func destroyFromTemplate(client *openstack.Openstack, template *CreateTemplate) error {

	for _, server := range template.Servers {
		if err := deleteServer(client, server, true); err != nil {
			return err
		}
	}
	for _, network := range template.Networks {
		if err := deleteNetwork(client, network); err != nil {
			return err
		}
	}
	for _, flavor := range template.Flavors {
		if err := deleteFlavor(client, flavor); err != nil {
			return err
		}
	}
	return nil
}

var UndefineCmd = &cobra.Command{
	Use:     "undefine <file>",
	Short:   i18n.T("undefineResourcesFromTempFile"),
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"delete"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		utility.LogError(err, "load template file failed", true)
		client := common.DefaultClient()
		destroyFromTemplate(client, createTemplate)
	},
}
