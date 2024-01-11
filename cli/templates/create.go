package templates

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/openstack/image"
)

func getFlavor(computeClient *compute.ComputeClientV2, template CreateTemplate) (*compute.Flavor, error) {
	newFlavor := compute.Flavor{
		Id:    template.Flavor.Id,
		Name:  template.Flavor.Name,
		Vcpus: template.Flavor.Vcpus,
		Ram:   template.Flavor.Ram,
	}
	f, _ := computeClient.FlavorShow(newFlavor.Id)
	if f == nil {
		logging.Info("creating flavor %s", newFlavor.Id)
		f, err := computeClient.FlavorCreate(newFlavor)
		common.LogError(err, "create flavor failed", true)
		logging.Info("creating flavor extra specs")
		if template.Flavor.ExtraSpecs != nil {
			_, err = computeClient.FlavorExtraSpecsCreate(f.Id, template.Flavor.ExtraSpecs)
			common.LogError(err, "create flavor extra specs failed", true)
		}
	}

	if template.Server.Flavor.Id != "" {
		logging.Info("find flavor %s", template.Server.Flavor.Id)
		return computeClient.FlavorShow(template.Server.Flavor.Id)
	} else if template.Server.Flavor.Name != "" {
		logging.Info("find flavor %s", template.Server.Flavor.Name)
		return computeClient.FlavorFoundByName(template.Server.Flavor.Name)
	} else {
		return nil, fmt.Errorf("flavor is empty")
	}
}
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

var CreateCmd = &cobra.Command{
	Use:   "create <file>",
	Short: "create from template file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		createTemplate, err := LoadCreateTemplate(args[0])
		common.LogError(err, "load template file failed", true)

		client := cli.GetClient()
		computClient := client.ComputeClient()
		imageClient := client.ImageClient()
		if createTemplate.Server.Name == "" {
			createTemplate.Server.Name = fmt.Sprintf(
				"%s%s", createTemplate.Default.ServerNamePrefix,
				time.Now().Format("2006-01-02_15:04:05"),
			)
		} else {
			logging.Info("find server %s", createTemplate.Server.Name)
			server, _ := client.ComputeClient().ServerFound(createTemplate.Server.Name)
			if server != nil {
				logging.Error("found server %s", server.Name)
				os.Exit(1)
			}
		}
		flavor, err := getFlavor(computClient, *createTemplate)
		common.LogError(err, "get flavor failed", true)

		img, err := getImage(imageClient, createTemplate.Server.Image)
		common.LogError(err, "get image failed", true)

		serverOption := compute.ServerOpt{
			Name:             createTemplate.Server.Name,
			Flavor:           flavor.Id,
			AvailabilityZone: createTemplate.Server.AvailabilityZone,
			MinCount:         createTemplate.Server.Min,
			MaxCount:         createTemplate.Server.Max,
		}
		if createTemplate.Server.VolumeBoot {
			serverOption.BlockDeviceMappingV2 = []compute.BlockDeviceMappingV2{
				{
					UUID:               img.Id,
					VolumeSize:         createTemplate.Server.VolumeSize,
					SourceType:         "image",
					DestinationType:    "volume",
					DeleteOnTemination: true,
				},
			}
			if createTemplate.Server.VolumeType != "" {
				serverOption.BlockDeviceMappingV2[0].VolumeType = createTemplate.Server.VolumeType
			}
		} else {
			serverOption.Image = img.Id
		}
		if createTemplate.Server.Network != "" {
			serverOption.Networks = []compute.ServerOptNetwork{
				{UUID: createTemplate.Server.Network},
			}
		}

		server, err := client.ComputeClient().ServerCreate(serverOption)
		common.LogError(err, "create server failed", true)
		logging.Info("creating server %s", serverOption.Name)
		server, err = client.ComputeClient().ServerShow(server.Id)
		common.LogError(err, "get server failed", true)
		cli.PrintServer(*server)
	},
}
