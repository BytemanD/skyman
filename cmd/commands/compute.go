package commands

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Server = &cobra.Command{Use: "server"}

var ServerList = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()

		query := url.Values{}
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")

		if name != "" {
			query.Set("name", name)
		}
		if host != "" {
			query.Set("host", host)
		}
		for _, status := range statusList {
			query.Add("status", status)
		}

		serversTable := ServersTable{Servers: computeClient.ServerListDetails(query)}
		serversTable.Print(long)
	},
}
var ServerShow = &cobra.Command{
	Use:   "show <name or id>",
	Short: "Show server details",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()
		nameOrId := args[0]
		server, _ := computeClient.ServerShow(nameOrId)
		serverTable := ServerTable{Server: server}
		serverTable.Print()
	},
}
var ServerCreate = &cobra.Command{
	Use:   "create",
	Short: "Create server(s)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		CLIENT := getComputeClient()
		flavor, _ := cmd.Flags().GetString("flavor")
		image, _ := cmd.Flags().GetString("image")
		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetInt("volume-size")
		az, _ := cmd.Flags().GetString("az")

		if flavor == "" {
			flavor = common.CONF.Server.Flavor
		}
		if image == "" {
			image = common.CONF.Server.Image
		}
		if volumeSize <= 0 {
			volumeSize = common.CONF.Server.VolumeSize
		}
		if !volumeBoot {
			volumeBoot = common.CONF.Server.VolumeBoot
		}
		if az == "" {
			az = common.CONF.Server.AvailabilityZone
		}

		createOption := compute.ServerOpt{
			Name: name, Flavor: flavor, Image: image,
			AvailabilityZone: az,
		}

		if !volumeBoot {
			createOption.Image = image
		} else {
			createOption.BlockDeviceMappingV2 = []compute.BlockDeviceMappingV2{
				{
					UUID: image, VolumeSize: volumeSize,
					SourceType: "image", DestinationType: "volume",
					DeleteOnTemination: true,
				},
			}
		}
		server, err := CLIENT.ServerCreate(createOption)
		if err != nil {
			logging.Fatal("create server faield, %s", err)
		}
		server, _ = CLIENT.ServerShow(server.Id)
		table := ServerTable{Server: server}
		table.Print()
	},
}
var ServerSet = &cobra.Command{
	Use:   "set",
	Short: "Set server properties",
	Run: func(cmd *cobra.Command, args []string) {
		logging.Info("list servers")
	},
}
var ServerDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		computeClient := getComputeClient()
		for _, id := range args {
			err := computeClient.ServerDelete(id)
			if err != nil {
				logging.Error("Reqeust to delete server failed, %v", err)
			} else {
				fmt.Printf("Requested to delete server: %s\n", id)
			}
		}
	},
}

func init() {
	ServerList.Flags().StringP("name", "n", "", "Search by server name")
	ServerList.Flags().String("host", "", "Search by hostname")
	ServerList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerList.Flags().BoolP("long", "l", false, "List additional fields in output")

	ServerCreate.Flags().StringP("flavor", "f", "", "Create server with this flavor")
	ServerCreate.Flags().StringP("image", "i", "", "Create server with this image")
	ServerCreate.Flags().StringP("nic", "n", "",
		"Create a NIC on the server. NIC format:\n"+
			"net-id=<net-uuid>: attach NIC to network with this UUID\n"+
			"port-id=<port-uuid>: attach NIC to port with this UUID",
	)
	ServerCreate.Flags().Bool("volume-boot", false, "Boot with volume")
	ServerCreate.Flags().Int("volume-size", 1, "Volume size(GB)")
	ServerCreate.Flags().String("az", "", "Select an availability zone for the server.")

	Server.AddCommand(ServerList)
	Server.AddCommand(ServerShow)
	Server.AddCommand(ServerCreate)
	Server.AddCommand(ServerDelete)
	Server.AddCommand(ServerSet)
}
