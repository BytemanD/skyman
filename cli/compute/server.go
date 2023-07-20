package compute

import (
	"fmt"
	"net/url"
	"time"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
	imageLib "github.com/BytemanD/stackcrud/openstack/image"
)

var Server = &cobra.Command{Use: "server"}

var serverList = &cobra.Command{
	Use:   "list",
	Short: "List servers",
	Run: func(cmd *cobra.Command, _ []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}

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
var serverShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show server details",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		nameOrId := args[0]
		server, err := client.Compute.ServerShow(nameOrId)
		if err != nil {
			servers := client.Compute.ServerListDetailsByName(nameOrId)
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
var serverCreate = &cobra.Command{
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
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
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
		server, err := client.Compute.ServerCreate(createOption)
		if err != nil {
			logging.Fatal("create server faield, %s", err)
		}
		server, _ = client.Compute.ServerShow(server.Id)
		server.Print()
	},
}
var serverSet = &cobra.Command{
	Use:   "set",
	Short: "Set server properties",
	Run: func(_ *cobra.Command, _ []string) {
		logging.Info("list servers")
	},
}
var serverDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}

		for _, id := range args {
			err := client.Compute.ServerDelete(id)
			if err != nil {
				logging.Error("Reqeust to delete server failed, %v", err)
			} else {
				fmt.Printf("Requested to delete server: %s\n", id)
			}
		}
	},
}
var serverPrune = &cobra.Command{
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
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		client.Compute.ServerPrune(query, yes, wait)
	},
}

func init() {
	// Server list flags
	serverList.Flags().StringP("name", "n", "", "Search by server name")
	serverList.Flags().String("host", "", "Search by hostname")
	serverList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverList.Flags().BoolP("long", "l", false, "List additional fields in output")
	serverList.Flags().BoolP("verbose", "v", false, "List verbose fields in output")
	// Server create flags
	serverCreate.Flags().StringP("flavor", "f", "", "Create server with this flavor")
	serverCreate.Flags().StringP("image", "i", "", "Create server with this image")
	serverCreate.Flags().StringP("nic", "n", "",
		"Create a NIC on the server. NIC format:\n"+
			"net-id=<net-uuid>: attach NIC to network with this UUID\n"+
			"port-id=<port-uuid>: attach NIC to port with this UUID")
	serverCreate.Flags().Bool("volume-boot", false, "Boot with volume")
	serverCreate.Flags().Uint16("volume-size", 0, "Volume size(GB)")
	serverCreate.Flags().String("az", "", "Select an availability zone for the server.")
	serverCreate.Flags().Uint16("min", 1, "Minimum number of servers to launch.")
	serverCreate.Flags().Uint16("max", 1, "Maximum number of servers to launch.")

	// Server prune flags
	serverPrune.Flags().StringP("name", "n", "", "Search by server name")
	serverPrune.Flags().String("host", "", "Search by hostname")
	serverPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverPrune.Flags().BoolP("wait", "w", false, "等待虚拟删除完成")
	serverPrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	Server.AddCommand(
		serverList, serverShow, serverCreate, serverDelete, serverSet,
		serverPrune)
}
