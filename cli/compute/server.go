package compute

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
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
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()
		query := url.Values{}
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")
		all, _ := cmd.Flags().GetBool("all")
		dsc, _ := cmd.Flags().GetBool("dsc")

		if name != "" {
			query.Set("name", name)
		}
		if host != "" {
			query.Set("host", host)
		}
		if all {
			query.Set("all_tenants", strconv.FormatBool(all))
		}
		for _, status := range statusList {
			query.Add("status", status)
		}

		long, _ := cmd.Flags().GetBool("long")
		verbose, _ := cmd.Flags().GetBool("verbose")
		servers, err := client.Compute.ServerListDetails(query)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Name", "Status", "TaskState", "PowerState", "Addresses"},
			LongHeaders: []string{
				"AZ", "Host", "InstanceName", "Flavor:Name"},
			HeaderLabel: map[string]string{"Addresses": "Networks"},
			SortBy: []table.SortBy{
				{Name: "Name", Mode: table.Asc},
				{Name: "Id", Mode: table.Asc},
			},
			Slots: map[string]func(item interface{}) interface{}{
				"PowerState": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return cli.BaseColorFormatter.Format(p.GetPowerState())
				},
				"Addresses": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return strings.Join(p.GetNetworks(), "\n")
				},
				"Flavor:Name": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.OriginalName
				},
				"Flavor:ram": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.Ram
				},
				"Flavor:vcpus": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.Vcpus
				},
				"Image": func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Image.Name
				},
				"Status": func(item interface{}) interface{} {
					p, _ := item.(compute.Server)
					return cli.BaseColorFormatter.Format(p.Status)
				},
			},
		}
		if dsc {
			dataTable.SortBy[0].Mode = table.Dsc
			dataTable.SortBy[1].Mode = table.Dsc
		}
		if long {
			dataTable.StyleSeparateRows = true
		}
		if verbose {
			dataTable.LongHeaders = append(dataTable.LongHeaders,
				"Flavor:ram", "Flavor:vcpus", "Image")
		}
		dataTable.AddItems(servers)
		dataTable.Print(long)
	},
}
var serverShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show server details",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()

		nameOrId := args[0]
		server, err := client.Compute.ServerShow(nameOrId)
		if err != nil {
			servers, err := client.Compute.ServerListDetailsByName(nameOrId)
			if len(servers) > 1 {
				fmt.Printf("Found multy severs named %s\n", nameOrId)
				os.Exit(1)
			} else if len(servers) == 1 {
				server = &servers[0]
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		printServer(*server)
	},
}
var serverCreate = &cobra.Command{
	Use:   "create [server name]",
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
		client := cli.GetClient()

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
			fmt.Println(err)
			os.Exit(1)
		}
		server, err = client.Compute.ServerShow(server.Id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		printServer(*server)
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
		client := cli.GetClient()

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
		client := cli.GetClient()

		client.Compute.ServerPrune(query, yes, wait)
	},
}
var serverStop = &cobra.Command{
	Use:   "stop <server> [<server> ...]",
	Short: "Stop server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, id := range args {
			err := client.Compute.ServerStop(id)
			if err != nil {
				logging.Error("Reqeust to stop server failed, %v", err)
			} else {
				fmt.Printf("Requested to stop server: %s\n", id)
			}
		}
	},
}
var serverStart = &cobra.Command{
	Use:   "start <server> [<server> ...]",
	Short: "Start server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := cli.GetClient()
		for _, id := range args {
			err := client.Compute.ServerStart(id)
			if err != nil {
				logging.Error("Reqeust to start server failed, %v", err)
			} else {
				fmt.Printf("Requested to start server: %s\n", id)
			}
		}
	},
}
var serverReboot = &cobra.Command{
	Use:   "reboot <server> [<server> ...]",
	Short: "Reboot server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		hard, _ := cmd.Flags().GetBool("hard")
		for _, id := range args {
			err := client.Compute.ServerReboot(id, hard)
			if err != nil {
				logging.Error("Reqeust to reboot server failed, %v", err)
			} else {
				fmt.Printf("Requested to reboot server: %s\n", id)
			}
		}
	},
}
var serverPause = &cobra.Command{
	Use:   "pause <server> [<server> ...]",
	Short: "Pause server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerPause(id)
			if err != nil {
				logging.Error("Reqeust to pause server failed, %v", err)
			} else {
				fmt.Printf("Requested to pause server: %s\n", id)
			}
		}
	},
}
var serverUnpause = &cobra.Command{
	Use:   "unpause <server> [<server> ...]",
	Short: "Unpause server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerUnpause(id)
			if err != nil {
				logging.Error("Reqeust to unpause server failed, %v", err)
			} else {
				fmt.Printf("Requested to unpause server: %s\n", id)
			}
		}
	},
}
var serverShelve = &cobra.Command{
	Use:   "shelve <server> [<server> ...]",
	Short: "Shelve server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerShelve(id)
			if err != nil {
				logging.Error("Reqeust to shelve server failed, %v", err)
			} else {
				fmt.Printf("Requested to shelve server: %s\n", id)
			}
		}
	},
}
var serverUnshelve = &cobra.Command{
	Use:   "unshelve <server> [<server> ...]",
	Short: "Unshelve server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerShelve(id)
			if err != nil {
				logging.Error("Reqeust to unshelve server failed, %v", err)
			} else {
				fmt.Printf("Requested to unshelve server: %s\n", id)
			}
		}
	},
}
var serverSuspend = &cobra.Command{
	Use:   "suspend <server> [<server> ...]",
	Short: "Suspend server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerSuspend(id)
			if err != nil {
				logging.Error("Reqeust to susppend server failed, %v", err)
			} else {
				fmt.Printf("Requested to susppend server: %s\n", id)
			}
		}
	},
}
var serverResume = &cobra.Command{
	Use:   "resume <server> [<server> ...]",
	Short: "Resume server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		for _, id := range args {
			err := client.Compute.ServerResume(id)
			if err != nil {
				logging.Error("Reqeust to resume server failed, %v", err)
			} else {
				fmt.Printf("Requested to resume server: %s\n", id)
			}
		}
	},
}
var serverResize = &cobra.Command{
	Use:   "resize <server> <flavor>",
	Short: "Resize server",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		err := client.Compute.ServerResize(args[0], args[1])
		if err != nil {
			logging.Error("Reqeust to resize server failed, %v", err)
		} else {
			fmt.Printf("Requested to resize server: %s\n", args[0])
		}
	},
}
var serverMigrate = &cobra.Command{
	Use:   "migrate <server>",
	Short: "Migrate server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		live, _ := cmd.Flags().GetBool("live")
		host, _ := cmd.Flags().GetString("host")
		blockMigrate, _ := cmd.Flags().GetBool("block-migrate")
		var err error
		if live {
			if host == "" {
				err = client.Compute.ServerLiveMigrate(args[0], blockMigrate)
			} else {
				err = client.Compute.ServerLiveMigrateTo(args[0], blockMigrate, host)
			}
		} else {
			if host == "" {
				err = client.Compute.ServerMigrate(args[0])
			} else {
				err = client.Compute.ServerMigrateTo(args[0], host)
			}
		}

		if err != nil {
			logging.Error("Reqeust to migrate server failed, %v", err)
		} else {
			fmt.Printf("Requested to migrate server: %s\n", args[0])
		}
	},
}

var serverRebuild = &cobra.Command{
	Use:   "rebuild <server>",
	Short: "Rebuild server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		err := client.Compute.ServerRebuild(args[0])
		if err != nil {
			logging.Error("Reqeust to rebuild server failed, %v", err)
		} else {
			fmt.Printf("Requested to rebuild server: %s\n", args[0])
		}
	},
}

func init() {
	// Server list flags
	serverList.Flags().StringP("name", "n", "", "Search by server name")
	serverList.Flags().String("host", "", "Search by hostname")
	serverList.Flags().BoolP("all", "a", false, "Display servers from all tenants")
	serverList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverList.Flags().BoolP("long", "l", false, "List additional fields in output")
	serverList.Flags().BoolP("verbose", "v", false, "List verbose fields in output")
	serverList.Flags().Bool("dsc", false, "Sort name by dsc")
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

	// server reboot flags

	serverReboot.Flags().Bool("hard", false, "Perform a hard reboot")

	// server migrate flags
	serverMigrate.Flags().Bool("live", false, "Migrate running server.")
	serverMigrate.Flags().String("host", "", "Destination host name.")
	serverMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration.")

	// Server prune flags
	serverPrune.Flags().StringP("name", "n", "", "Search by server name")
	serverPrune.Flags().String("host", "", "Search by hostname")
	serverPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverPrune.Flags().BoolP("wait", "w", false, "等待虚拟删除完成")
	serverPrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	Server.AddCommand(
		serverList, serverShow, serverCreate, serverDelete, serverPrune,
		serverSet, serverStop, serverStart, serverReboot,
		serverPause, serverUnpause, serverShelve, serverUnshelve,
		serverSuspend, serverResume, serverResize, serverRebuild,
		serverMigrate)
}
