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
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")
		if min < 0 || max < 0 {
			return fmt.Errorf("Invalid flags: expect min and max > 0, got %d, %d", min, max)
		}
		if min > max {
			return fmt.Errorf("Invalid flags: expect min <= max, got: %d > %d", min, max)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		CLIENT := getComputeClient()
		flavor, _ := cmd.Flags().GetString("flavor")
		image, _ := cmd.Flags().GetString("image")
		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetInt("volume-size")
		az, _ := cmd.Flags().GetString("az")

		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")

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
			Name:             name,
			Flavor:           flavor,
			Image:            image,
			AvailabilityZone: az,
			MinCount:         min,
			MaxCount:         max,
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
var ServerPrune = &cobra.Command{
	Use:   "prune",
	Short: "Prune server(s)",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		yes, _ := cmd.Flags().GetBool("yes")
		wait, _ := cmd.Flags().GetBool("wait")
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")

		query := url.Values{}
		if name != "" {
			query.Set("name", name)
		}
		if host != "" {
			query.Set("host", host)
		}
		for _, status := range statusList {
			query.Add("status", status)
		}

		computeClient := getComputeClient()

		computeClient.ServerPrune(query, yes, wait)
	},
}

func init() {
	// Server list flags
	ServerList.Flags().StringP("name", "n", "", "Search by server name")
	ServerList.Flags().String("host", "", "Search by hostname")
	ServerList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerList.Flags().BoolP("long", "l", false, "List additional fields in output")
	// Server create flags
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
	ServerCreate.Flags().Uint16("min", 1, "Minimum number of servers to launch.")
	ServerCreate.Flags().Uint16("max", 1, "Maximum number of servers to launch.")

	// Server prune flags
	ServerPrune.Flags().StringP("name", "n", "", "Search by server name")
	ServerPrune.Flags().String("host", "", "Search by hostname")
	ServerPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerPrune.Flags().BoolP("wait", "w", false, "等待虚拟删除完成")
	ServerPrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	Server.AddCommand(ServerList)
	Server.AddCommand(ServerShow)
	Server.AddCommand(ServerCreate)
	Server.AddCommand(ServerDelete)
	Server.AddCommand(ServerSet)
	Server.AddCommand(ServerPrune)
}
