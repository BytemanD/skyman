package templates

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func deleteFlavor(client *openstack.Openstack, flavor Flavor) error {
	var f *nova.Flavor
	var err error
	if flavor.Id != "" {
		f, err = client.NovaV2().Flavor().Show(flavor.Id)
		if err != nil {
			logging.Warning("get flavor %s failed: %s", flavor.Id, err)
			return nil
		}
	} else if flavor.Name != "" {
		f, err = client.NovaV2().Flavor().Found(flavor.Name)
		if err != nil {
			logging.Warning("get flavor %s failed, %s", flavor.Name, err)
			return nil
		}
	}
	logging.Info("deleting flavor %s", f.Id)
	return client.NovaV2().Flavor().Delete(f.Id)
	// utility.LogError(err, fmt.Sprintf("delete flavor %s failed", f.Id), false)
}
func deleteNetwork(client *openstack.Openstack, network Network) error {
	net, err := client.NeutronV2().Network().Found(network.Name)
	if err != nil {
		logging.Warning("get network %s failed: %s", network.Name, err)
		return nil
	}
	logging.Info("deleting network %s", network.Name)
	return client.NeutronV2().Network().Delete(net.Id)

	// utility.LogError(err, fmt.Sprintf("delete network %s failed: %s", network.Name, err), false)
}

func deleteServer(client *openstack.Openstack, server Server, watch bool) error {
	s, err := client.NovaV2().Server().Found(server.Name)
	if err != nil {
		logging.Warning("get server %s failed, %s", server.Name, err)
		return nil
	}
	err = client.NovaV2().Server().Delete(s.Id)
	if err != nil {
		// utility.LogError(err, fmt.Sprintf("delete server %s failed", s.Name), false)
		return nil
	}

	logging.Info("[%s] deleting (%s)", s.Id, s.Name)
	if watch {
		return client.NovaV2().Server().WaitDeleted(s.Id)
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
		client := openstack.DefaultClient()
		destroyFromTemplate(client, createTemplate)
	},
}
