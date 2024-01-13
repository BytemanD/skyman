package templates

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	openstackCommon "github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/openstack/image"
)

func getImage(imageClient *image.ImageClientV2, resource BaseResource) (*image.Image, error) {
	if resource.Id != "" {
		logging.Info("find image %s", resource.Id)
		return imageClient.ImageShow(resource.Id)
	} else if resource.Name != "" {
		logging.Info("find image %s", resource.Name)
		return imageClient.ImageFountByName(resource.Name)
	} else {
		return nil, fmt.Errorf("image is empty")
	}
}
func createFlavor(client *openstack.OpenstackClient, flavor Flavor) {
	computeClient := client.ComputeClient()
	f, _ := computeClient.FlavorShow(flavor.Id)
	if f != nil {
		logging.Warning("network %s exists", flavor.Id)
		return
	}
	newFlavor := compute.Flavor{
		Id:    flavor.Id,
		Name:  flavor.Name,
		Vcpus: flavor.Vcpus,
		Ram:   flavor.Ram,
	}
	logging.Info("creating flavor %s", newFlavor.Id)
	f, err := computeClient.FlavorCreate(newFlavor)
	common.LogError(err, "create flavor failed", true)
	if flavor.ExtraSpecs != nil {
		logging.Info("creating flavor extra specs")
		_, err = computeClient.FlavorExtraSpecsCreate(f.Id, flavor.ExtraSpecs)
		common.LogError(err, "create flavor extra specs failed", true)
	}
}
func createNetwork(client *openstack.OpenstackClient, network Network) {
	networkClient := client.NetworkingClient()
	_, err := networkClient.NetworkFound(network.Name)
	if err == nil {
		logging.Warning("network %s exists", network.Name)
		return
	}
	netParams := map[string]interface{}{
		"name": network.Name,
	}
	logging.Info("creating network %s", network.Name)
	net, err := networkClient.NetworkCreate(netParams)
	common.LogError(err, fmt.Sprintf("create network %s failed", network.Name), true)
	for _, subnet := range network.Subnets {
		if subnet.IpVersion == 0 {
			subnet.IpVersion = 4
		}
		subnetParams := map[string]interface{}{
			"name":       subnet.Name,
			"network_id": net.Id,
			"cidr":       subnet.Cidr,
			"ip_version": subnet.IpVersion,
		}
		logging.Info("creating subnet %s (cidr: %s)", subnet.Name, subnet.Cidr)
		_, err2 := networkClient.SubnetCreate(subnetParams)
		common.LogError(err2, fmt.Sprintf("create subnet %s failed", subnet.Name), true)
	}
}

func createServer(client *openstack.OpenstackClient, server Server, watch bool) (*compute.Server, error) {
	computeClient := client.ComputeClient()
	imageClient := client.ImageClient()
	networkClient := client.NetworkingClient()

	s, _ := client.ComputeClient().ServerFound(server.Name)
	if s != nil {
		logging.Warning("server %s exists", s.Name)
		return s, nil
	}
	serverOption := compute.ServerOpt{
		Name:             server.Name,
		AvailabilityZone: server.AvailabilityZone,
		MinCount:         server.Min,
		MaxCount:         server.Max,
	}
	var flavor *compute.Flavor
	var err error
	if server.Flavor.Id == "" && server.Flavor.Name == "" {
		return nil, fmt.Errorf("flavor is empty")
	}

	if server.Flavor.Id != "" {
		logging.Info("find flavor %s", server.Flavor.Id)
		flavor, err = computeClient.FlavorShow(server.Flavor.Id)
	} else if server.Flavor.Name != "" {
		logging.Info("find flavor %s", server.Flavor.Name)
		flavor, err = computeClient.FlavorFoundByName(server.Flavor.Name)
	}
	common.LogError(err, "get flavor failed", true)
	serverOption.Flavor = flavor.Id

	if server.Image.Id != "" || server.Image.Name != "" {
		img, err := getImage(imageClient, server.Image)
		common.LogError(err, "get image failed", true)
		serverOption.Image = img.Id
	}

	if len(server.BlockDeviceMappingV2) > 0 {
		serverOption.BlockDeviceMappingV2 = []compute.BlockDeviceMappingV2{}
		for _, bdm := range server.BlockDeviceMappingV2 {
			serverOption.BlockDeviceMappingV2 = append(serverOption.BlockDeviceMappingV2,
				compute.BlockDeviceMappingV2{
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
		networks := []compute.ServerOptNetwork{}
		for _, nic := range server.Networks {
			if nic.UUID != "" {
				networks = append(networks, compute.ServerOptNetwork{UUID: nic.UUID})
			} else if nic.Port != "" {
				networks = append(networks, compute.ServerOptNetwork{Port: nic.Port})
			} else if nic.Name != "" {
				network, err := networkClient.NetworkFound(nic.Name)
				common.LogError(err, "found network failed", true)
				networks = append(networks, compute.ServerOptNetwork{UUID: network.Id})
			}
		}
		serverOption.Networks = networks
	}
	if server.UserData != "" {
		content, err := openstackCommon.LoadUserData(server.UserData)
		common.LogError(err, "read user data failed", true)
		serverOption.UserData = content
	}
	s, err = computeClient.ServerCreate(serverOption)
	common.LogError(err, "create server failed", true)
	logging.Info("creating server %s", serverOption.Name)
	if watch {
		client.WaitServerCreated(s.Id)
	}
	return s, nil
}

var CreateCmd = &cobra.Command{
	Use:   "create <file>",
	Short: "create from template file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		watch, _ := cmd.Flags().GetBool("watch")
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		common.LogError(err, "load template file failed", true)

		for _, server := range createTemplate.Servers {
			if server.Name == "" {
				logging.Fatal("invalid config, server name is empty")
			}
			if server.Flavor.Id == "" && server.Flavor.Name == "" {
				logging.Fatal("invalid config, server flavor is empty")
			}
			if server.Image.Id == "" && server.Image.Name == "" && len(server.BlockDeviceMappingV2) == 0 {
				logging.Fatal("invalid config, server image is empty")
			}
		}

		client := cli.GetClient()

		for _, flavor := range createTemplate.Flavors {
			createFlavor(client, flavor)
		}
		for _, network := range createTemplate.Networks {
			createNetwork(client, network)
		}

		for _, server := range createTemplate.Servers {
			_, err := createServer(client, server, watch)
			common.LogError(err, "create server failed", true)
		}
	},
}

func init() {
	CreateCmd.Flags().Bool("watch", false, "watch the resource progress until it completes.")
}
