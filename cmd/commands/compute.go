package commands

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack"
	"github.com/BytemanD/stackcrud/openstack/compute"
	imageLib "github.com/BytemanD/stackcrud/openstack/image"
)

var Server = &cobra.Command{Use: "server"}
var Compute = &cobra.Command{Use: "compute"}
var computeService = &cobra.Command{Use: "service"}

var ServerList = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	Run: func(cmd *cobra.Command, _ []string) {
		authClient := getAuthClient()
		client, err := openstack.GetClientWithAuthToken(authClient)
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		client.Compute.UpdateVersion()
		query := url.Values{}
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

		long, _ := cmd.Flags().GetBool("long")
		verbose, _ := cmd.Flags().GetBool("verbose")
		servers := client.Compute.ServerListDetails(query)

		imageMap := map[string]imageLib.Image{}
		if long && verbose {
			for i, server := range servers {
				if _, ok := imageMap[server.Image.Id]; !ok {
					image, err := client.Image.ImageShow(server.Image.Id)
					if err != nil {
						logging.Warning("get image %s faield, %s", server.Image.Id, err)
					} else {
						imageMap[server.Image.Id] = *image
					}
				}
				servers[i].Image.Name = imageMap[server.Image.Id].Name
			}
		}
		servers.Print(long, verbose)
	},
}
var ServerShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show server details",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		computeClient := getComputeClient()
		nameOrId := args[0]
		server, err := computeClient.ServerShow(nameOrId)
		if err != nil {
			servers := computeClient.ServerListDetailsByName(nameOrId)
			if len(servers) > 1 {
				fmt.Printf("Found multy severs named %s\n", nameOrId)
			} else if len(servers) == 1 {
				server = &servers[0]
			} else {
				fmt.Println(err)
			}
		}
		if server != nil {
			server.Print()
		}
	},
}
var ServerCreate = &cobra.Command{
	Use:   "create",
	Short: "Create server(s)",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
			return err
		}
		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")
		if min > max {
			return fmt.Errorf("invalid flags: expect min <= max, got: %d > %d", min, max)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var name string
		if len(args) == 1 {
			name = args[0]
		} else {
			name = fmt.Sprintf(
				"%s%s", common.CONF.Server.NamePrefix,
				time.Now().Format("2006-01-02_15:04:05"),
			)

		}
		CLIENT := getComputeClient()
		flavor, _ := cmd.Flags().GetString("flavor")
		image, _ := cmd.Flags().GetString("image")
		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetUint16("volume-size")
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
		server.Print()
	},
}
var ServerSet = &cobra.Command{
	Use:   "set",
	Short: "Set server properties",
	Run: func(_ *cobra.Command, _ []string) {
		logging.Info("list servers")
	},
}
var ServerDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
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
	Run: func(cmd *cobra.Command, _ []string) {
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
var csList = &cobra.Command{
	Use:   "list",
	Short: "List compute services",
	Run: func(cmd *cobra.Command, _ []string) {
		authClient := getAuthClient()
		client, err := openstack.GetClientWithAuthToken(authClient)
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		client.Compute.UpdateVersion()
		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")

		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}

		services := client.Compute.ServiceList(query)
		services.PrintTable(long)
	},
}

func init() {
	// Server list flags
	ServerList.Flags().StringP("name", "n", "", "Search by server name")
	ServerList.Flags().String("host", "", "Search by hostname")
	ServerList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerList.Flags().BoolP("long", "l", false, "List additional fields in output")
	ServerList.Flags().BoolP("verbose", "v", false, "List verbose fields in output")
	// Server create flags
	ServerCreate.Flags().StringP("flavor", "f", "", "Create server with this flavor")
	ServerCreate.Flags().StringP("image", "i", "", "Create server with this image")
	ServerCreate.Flags().StringP("nic", "n", "",
		"Create a NIC on the server. NIC format:\n"+
			"net-id=<net-uuid>: attach NIC to network with this UUID\n"+
			"port-id=<port-uuid>: attach NIC to port with this UUID",
	)
	ServerCreate.Flags().Bool("volume-boot", false, "Boot with volume")
	ServerCreate.Flags().Uint16("volume-size", 0, "Volume size(GB)")
	ServerCreate.Flags().String("az", "", "Select an availability zone for the server.")
	ServerCreate.Flags().Uint16("min", 1, "Minimum number of servers to launch.")
	ServerCreate.Flags().Uint16("max", 1, "Maximum number of servers to launch.")

	// Server prune flags
	ServerPrune.Flags().StringP("name", "n", "", "Search by server name")
	ServerPrune.Flags().String("host", "", "Search by hostname")
	ServerPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	ServerPrune.Flags().BoolP("wait", "w", false, "等待虚拟删除完成")
	ServerPrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	// compute service
	csList.Flags().String("binary", "", "Search by binary")
	csList.Flags().String("host", "", "Search by hostname")
	csList.Flags().StringArrayP("state", "s", nil, "Search by server status")
	csList.Flags().BoolP("long", "l", false, "List additional fields in output")
	computeService.AddCommand(csList)

	Server.AddCommand(
		ServerList, ServerShow, ServerCreate, ServerDelete, ServerSet,
		ServerPrune)

	Compute.AddCommand(computeService)
}
