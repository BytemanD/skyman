package templates

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func getImage(client *openstack.Openstack, resource BaseResource) (*glance.Image, error) {
	if resource.Id != "" {
		console.Info("find image %s", resource.Id)
		return client.GlanceV2().GetImage(resource.Id)
	} else if resource.Name != "" {
		console.Info("find image %s", resource.Name)
		return client.GlanceV2().FindImage(resource.Name)
	} else {
		return nil, fmt.Errorf("image is empty")
	}
}
func createFlavor(client *openstack.Openstack, flavor Flavor) {
	computeClient := client.NovaV2()
	f, _ := computeClient.GetFlavor(flavor.Id)
	if f != nil {
		console.Warn("network %s exists", flavor.Id)
		return
	}
	newFlavor := nova.Flavor{
		Id:    flavor.Id,
		Name:  flavor.Name,
		Vcpus: flavor.Vcpus,
		Ram:   flavor.Ram,
	}
	console.Info("creating flavor %s", newFlavor.Id)
	f, err := computeClient.CreateFlavor(newFlavor)
	utility.LogError(err, "create flavor failed", true)
	if flavor.ExtraSpecs != nil {
		console.Info("creating flavor extra specs")
		_, err = computeClient.SetFlavorExtraSpecs(f.Id, flavor.ExtraSpecs)
		utility.LogError(err, "create flavor extra specs failed", true)
	}
}
func createNetwork(client *openstack.Openstack, network Network) {
	networkClient := client.NeutronV2()
	_, err := networkClient.FindNetwork(network.Name)
	if err == nil {
		console.Warn("network %s exists", network.Name)
		return
	}
	netParams := map[string]any{
		"name": network.Name,
	}
	console.Info("creating network %s", network.Name)
	net, err := networkClient.CreateNetwork(netParams)
	utility.LogError(err, fmt.Sprintf("create network %s failed", network.Name), true)
	for _, subnet := range network.Subnets {
		if subnet.IpVersion == 0 {
			subnet.IpVersion = 4
		}
		subnetParams := map[string]any{
			"name":       subnet.Name,
			"network_id": net.Id,
			"cidr":       subnet.Cidr,
			"ip_version": subnet.IpVersion,
		}
		console.Info("creating subnet %s (cidr: %s)", subnet.Name, subnet.Cidr)
		_, err2 := networkClient.CreateSubnet(subnetParams)
		utility.LogError(err2, fmt.Sprintf("create subnet %s failed", subnet.Name), true)
	}
}

func createServer(client *openstack.Openstack, server Server, watch bool) (*nova.Server, error) {
	computeClient := client.NovaV2()
	networkClient := client.NeutronV2()

	s, _ := client.NovaV2().FindServer(server.Name)
	if s != nil {
		console.Warn("server %s exists (%s)", s.Name, s.Status)
		return s, nil
	}
	serverOption := nova.ServerOpt{
		Name:             server.Name,
		AvailabilityZone: server.AvailabilityZone,
		MinCount:         server.Min,
		MaxCount:         server.Max,
	}
	for _, sg := range server.SecurityGroups {
		serverOption.SecurityGroups = append(
			serverOption.SecurityGroups,
			neutron.SecurityGroup{
				Resource: model.Resource{Name: sg.Name},
			})
	}

	var flavor *nova.Flavor
	var err error
	if server.Flavor.Id == "" && server.Flavor.Name == "" {
		return nil, fmt.Errorf("flavor is empty")
	}

	if server.Flavor.Id != "" {
		console.Info("find flavor %s", server.Flavor.Id)
		flavor, err = client.NovaV2().GetFlavor(server.Flavor.Id)
	} else if server.Flavor.Name != "" {
		console.Info("find flavor %s", server.Flavor.Name)
		flavor, err = client.NovaV2().FindFlavor(server.Flavor.Name)
	}
	utility.LogError(err, "get flavor failed", true)
	serverOption.Flavor = flavor.Id

	if server.Image.Id != "" || server.Image.Name != "" {
		img, err := getImage(client, server.Image)
		utility.LogError(err, "get image failed", true)
		serverOption.Image = img.Id
	}

	if len(server.BlockDeviceMappingV2) > 0 {
		serverOption.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{}
		for _, bdm := range server.BlockDeviceMappingV2 {
			serverOption.BlockDeviceMappingV2 = append(serverOption.BlockDeviceMappingV2,
				nova.BlockDeviceMappingV2{
					BootIndex:          bdm.BootIndex,
					UUID:               bdm.UUID,
					VolumeSize:         bdm.VolumeSize,
					VolumeType:         bdm.VolumeType,
					SourceType:         bdm.SourceType,
					DestinationType:    bdm.DestinationType,
					DeleteOnTemination: bdm.DeleteOnTermination,
				},
			)
		}
	}
	if len(server.Networks) > 0 {
		networks := []nova.ServerOptNetwork{}
		for _, nic := range server.Networks {
			if nic.UUID != "" {
				networks = append(networks, nova.ServerOptNetwork{UUID: nic.UUID})
			} else if nic.Port != "" {
				networks = append(networks, nova.ServerOptNetwork{Port: nic.Port})
			} else if nic.Name != "" {
				network, err := networkClient.FindNetwork(nic.Name)
				utility.LogError(err, "found network failed", true)
				networks = append(networks, nova.ServerOptNetwork{UUID: network.Id})
			}
		}
		serverOption.Networks = networks
	}
	if server.UserData != "" {
		serverOption.UserData = utility.EncodedUserdata(server.UserData)
	}
	s, err = computeClient.CreateServer(serverOption)
	utility.LogError(err, "create server failed", true)
	console.Info("creating server %s", serverOption.Name)
	if watch {
		computeClient.WaitServerStatus(s.Id, "ACTIVE", 2)
	}
	return s, nil
}

var DefineCmd = &cobra.Command{
	Use:     "define <file>",
	Short:   i18n.T("defineResourcesFromTempFile"),
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"create"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		utility.LogError(err, "load template file failed", true)

		replace, _ := cmd.Flags().GetBool("replace")
		for _, server := range createTemplate.Servers {
			if server.Name == "" {
				console.Fatal("invalid config, server name is empty")
			}
			if server.Flavor.Id == "" && server.Flavor.Name == "" {
				console.Fatal("invalid config, server flavor is empty")
			}
			if server.Image.Id == "" && server.Image.Name == "" && len(server.BlockDeviceMappingV2) == 0 {
				console.Fatal("invalid config, server image is empty")
			}
		}
		client := common.DefaultClient()
		if replace {
			console.Info("destroy resources")
			if err := destroyFromTemplate(client, createTemplate); err != nil {
				utility.LogError(err, "destroy resource failed", true)
			}
		}

		for _, flavor := range createTemplate.Flavors {
			createFlavor(client, flavor)
		}
		for _, network := range createTemplate.Networks {
			createNetwork(client, network)
		}

		for _, server := range createTemplate.Servers {
			_, err := createServer(client, server, true)
			utility.LogError(err, "create server failed", true)
		}
	},
}

func init() {
	DefineCmd.Flags().Bool("replace", false, "Replace resources")
}
