package templates

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/template"
	"github.com/BytemanD/skyman/openstack/compute"
)

var CreateCmd = &cobra.Command{
	Use:   "create <file>",
	Short: "create from template file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		createTemplate, err := template.LoadCreateTemplate(args[0])
		common.LogError(err, "load template file failed", true)

		if createTemplate.Server.Name == "" {
			createTemplate.Server.Name = fmt.Sprintf(
				"%s%s", createTemplate.Server.NamePrefix,
				time.Now().Format("2006-01-02_15:04:05"),
			)
		}
		serverOption := compute.ServerOpt{
			Name:             createTemplate.Server.Name,
			Flavor:           createTemplate.Server.Flavor,
			AvailabilityZone: createTemplate.Server.AvailabilityZone,
			MinCount:         createTemplate.Server.Min,
			MaxCount:         createTemplate.Server.Max,
		}
		if createTemplate.Server.VolumeBoot {
			serverOption.BlockDeviceMappingV2 = []compute.BlockDeviceMappingV2{
				{
					UUID:               createTemplate.Server.Image,
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
			serverOption.Image = createTemplate.Server.Image
		}
		if createTemplate.Server.Network != "" {
			serverOption.Networks = []compute.ServerOptNetwork{
				{UUID: createTemplate.Server.Network},
			}
		}
		client := cli.GetClient()
		server, err := client.ComputeClient().ServerCreate(serverOption)
		common.LogError(err, "create server failed", true)
		logging.Info("creating server %s", serverOption.Name)
		server, err = client.ComputeClient().ServerShow(server.Id)
		common.LogError(err, "get server failed", true)
		cli.PrintServer(*server)
	},
}
