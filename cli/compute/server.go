package compute

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/howeyc/gopass"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/BytemanD/easygo/pkg/fileutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	openstackCommon "github.com/BytemanD/stackcrud/openstack/common"
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
		watch, _ := cmd.Flags().GetBool("watch")
		watchInterval, _ := cmd.Flags().GetUint16("watch-interval")

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true}, {Name: "Status", AutoColor: true},
				{Name: "TaskState"},
				{Name: "PowerState", AutoColor: true, Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.GetPowerState()
				}},
				{Name: "Addresses", Text: "Networks", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return strings.Join(p.GetNetworks(), "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "AZ"}, {Name: "Host"}, {Name: "InstanceName"},
				{Name: "Flavor:name", Text: "Flavor:name", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.OriginalName
				}},
			},
		}
		if dsc {
			pt.ShortColumns[1].SortMode = table.Dsc
		}
		if long {
			pt.StyleSeparateRows = true
		}
		if verbose {
			pt.LongColumns = append(pt.LongColumns,
				common.Column{Name: "Flavor:ram", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.Ram
				}},
				common.Column{Name: "Flavor:vcpus", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Flavor.Vcpus
				}},
				common.Column{Name: "Image:ram", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Server)
					return p.Image.Name
				}},
			)
		}

		imageMap := map[string]imageLib.Image{}
		for {
			servers, err := client.Compute.ServerListDetails(query)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
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
			pt.CleanItems()
			pt.AddItems(servers)
			if watch {
				cmd := exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
				var timeNow = time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("Every %ds\tNow: %s\n", watchInterval, timeNow)
			}
			common.PrintPrettyTable(pt, long)
			if !watch {
				break
			}
			time.Sleep(time.Second * time.Duration(watchInterval))
		}
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

		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetUint16("volume-size")

		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")
		userDataFile, _ := cmd.Flags().GetString("user-data")

		createOption := compute.ServerOpt{
			Name:             name,
			Flavor:           common.CONF.Server.Flavor,
			AvailabilityZone: common.CONF.Server.AvailabilityZone,
			MinCount:         min,
			MaxCount:         max,
		}

		if userDataFile != "" {
			content, err := fileutils.ReadAll(userDataFile)
			if err != nil {
				logging.Fatal("read user data %s failed, %v", userDataFile, err)
			}
			encodedContent := base64.StdEncoding.EncodeToString([]byte(content))
			createOption.UserData = encodedContent
		}

		if volumeSize <= 0 {
			volumeSize = common.CONF.Server.VolumeSize
		}
		if volumeBoot || common.CONF.Server.VolumeBoot {
			createOption.BlockDeviceMappingV2 = []compute.BlockDeviceMappingV2{
				{
					UUID:       common.CONF.Server.Image,
					VolumeSize: volumeSize,
					SourceType: "image", DestinationType: "volume",
					DeleteOnTemination: true,
				},
			}
		} else {
			createOption.Image = common.CONF.Server.Image
		}
		if common.CONF.Server.Network != "" {
			createOption.Networks = []compute.ServerOptNetwork{
				{UUID: common.CONF.Server.Network},
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
var serverSet = &cobra.Command{Use: "set"}
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
		wait, _ := cmd.Flags().GetBool("wait")
		client := cli.GetClient()

		flavor, err := client.Compute.FlavorShow(args[1])
		if err != nil {
			logging.Fatal("Get flavor %s failed, %v", args[1], err)
		}

		err = client.Compute.ServerResize(args[0], args[1])
		if err != nil {
			logging.Fatal("Reqeust to resize server failed, %v", err)
		} else {
			fmt.Printf("Requested to resize server: %s\n", args[0])
		}

		if wait {
			for {
				server, err := client.Compute.ServerShow(args[0])
				if err != nil {
					logging.Fatal("Get server %s failed, %v", args[0], err)
				}
				logging.Info("实例状态: %s", server.AllStatus())
				if server.TaskState != "" || server.InResize() {
					time.Sleep(time.Second * 2)
					continue
				}
				if server.Flavor.OriginalName == flavor.Name {
					logging.Info("resize 成功")
				} else {
					logging.Fatal("resize 失败")
				}
				break
			}
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

var serverEvacuate = &cobra.Command{
	Use:   "evacuate <server>",
	Short: "Evacuate server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		password, _ := cmd.Flags().GetString("password")
		host, _ := cmd.Flags().GetString("host")
		force, _ := cmd.Flags().GetBool("force")

		err := client.Compute.ServerEvacuate(args[0], password, host, force)
		if err != nil {
			common.LogError(err, "Reqeust to evacuate server failed", true)
		} else {
			fmt.Printf("Requested to evacuate server: %s\n", args[0])
		}
	},
}
var serverSetPassword = &cobra.Command{
	Use:   "password <server>",
	Short: "set password for server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		user, _ := cmd.Flags().GetString("user")
		server, err := client.Compute.ServerShow(args[0])
		if err != nil {
			logging.Fatal("show server %s failed, %v", args[0], err)
		}
		var (
			newPasswd []byte
			again     []byte
		)
		for {
			fmt.Printf("New password: ")
			newPasswd, _ = gopass.GetPasswd()
			if string(newPasswd) == "" {
				logging.Error("Password is empty.")
				continue
			}

			fmt.Printf("Again: ")
			again, _ = gopass.GetPasswd()
			if string(again) == string(newPasswd) {
				break
			}
			logging.Fatal("Passwords do not match.")
		}
		err = client.Compute.ServerSetPassword(server.Id, string(newPasswd), user)
		if err != nil {
			if httpError, ok := err.(*openstackCommon.HttpError); ok {
				logging.Fatal("set password failed, %s, %s",
					httpError.Reason, httpError.Message)
			} else {
				logging.Fatal("set password failed, %v", err)
			}
		} else {
			fmt.Println("Reqeust to set password successs")
		}
	},
}
var serverRegion = &cobra.Command{Use: "region"}
var serverRegionLiveMigrate = &cobra.Command{
	Use:   "migrate <server> <dest region>",
	Short: "Migrate server to another region",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]
		destRegion := args[1]
		live, _ := cmd.Flags().GetBool("live")
		host, _ := cmd.Flags().GetString("host")

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if !live {
			logging.Fatal("Only support live migrate now, please use option --live")
		}
		client := cli.GetClient()
		var (
			migrateErr  error
			migrateResp compute.RegionMigrateResp
		)
		if live {
			if host == "" {
				resp, err := client.Compute.ServerRegionLiveMigrate(serverId, destRegion, dryRun)
				migrateResp = *resp
				migrateErr = err
			} else {
				resp, err := client.Compute.ServerRegionLiveMigrateTo(serverId, destRegion, dryRun, host)
				migrateResp = *resp
				migrateErr = err
			}
			if dryRun {
				table := common.PrettyItemTable{
					Item: migrateResp,
					ShortFields: []common.Column{
						{Name: "AllowLiveMigrate"}, {Name: "Reason"},
					},
				}
				common.PrintPrettyItemTable(table)
			}
		}
		if migrateErr != nil {
			logging.Fatal("Reqeust to migrate server failed, %v", migrateErr)
		} else {
			fmt.Printf("Requested to migrate server: %s\n", args[0])
		}
	},
}
var serverMigration = &cobra.Command{Use: "migration"}
var serverMigrationList = &cobra.Command{
	Use:   "list <server>",
	Short: "List server migrations",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverId := args[0]
		status, _ := cmd.Flags().GetString("status")

		latest, _ := cmd.Flags().GetBool("latest")
		long, _ := cmd.Flags().GetBool("long")
		watch, _ := cmd.Flags().GetBool("watch")
		watchInterval, _ := cmd.Flags().GetUint16("watch-interval")
		migrateType, _ := cmd.Flags().GetString("type")

		query := url.Values{}
		if status != "" {
			query.Set("status", status)
		}
		if latest {
			query.Set("latest", "true")
		}
		if migrateType != "" {
			query.Set("type", migrateType)
		}
		client := cli.GetClient()
		table := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Status", AutoColor: true},
				{Name: "SourceNode"}, {Name: "SourceCompute"},
				{Name: "DestNode"}, {Name: "DestCompute"},
			},
			LongColumns: []common.Column{
				{Name: "DestHost"}, {Name: "CreatedAt"}, {Name: "UpdatedAt"},
			},
		}
		for {
			migrations, err := client.Compute.ServerMigrationList(serverId, query)
			if err != nil {
				logging.Fatal("Reqeust to list server migration failed, %v", err)
			}
			table.CleanItems()
			table.AddItems(migrations)
			if watch {
				cmd := exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
				var timeNow = time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("Every %ds\tNow: %s\n", watchInterval, timeNow)
			}
			common.PrintPrettyTable(table, long)
			if !watch {
				break
			}
			time.Sleep(time.Second * time.Duration(watchInterval))
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
	serverList.Flags().Bool("watch", false, "List loop")
	serverList.Flags().Uint16P("watch-interval", "i", 2, "Loop interval")
	// Server create flags
	serverCreate.Flags().String("flavor", "", "Create server with this flavor")
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
	serverCreate.Flags().String("user-data", "", "user data file to pass to be exposed by the metadata server.")

	viper.BindPFlag("server.flavor", serverCreate.Flags().Lookup("flavor"))
	viper.BindPFlag("server.image", serverCreate.Flags().Lookup("image"))
	viper.BindPFlag("server.availabilityZone", serverCreate.Flags().Lookup("az"))

	// server reboot flags
	serverReboot.Flags().Bool("hard", false, "Perform a hard reboot")

	// server migrate flags
	serverMigrate.Flags().Bool("live", false, "Migrate running server.")
	serverMigrate.Flags().String("host", "", "Destination host name.")
	serverMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration.")

	// server resize flags
	serverResize.Flags().BoolP("wait", "w", false, "Wait server resize completed")

	// server evacuate flags
	serverEvacuate.Flags().Bool("force", false, "Force to not verify the scheduler if a host is provided.")
	serverEvacuate.Flags().String("host", "", "Destination host name.")
	serverEvacuate.Flags().String("password", "", "Set the provided admin password on the evacuated server.")

	// server region migrate flags
	serverRegionLiveMigrate.Flags().Bool("live", false, "Migrate running server.")
	serverRegionLiveMigrate.Flags().String("host", "", "Destination host name.")
	// serverRegionLiveMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration.")
	serverRegionLiveMigrate.Flags().Bool("dry-run", false, "True in case of dry run.")
	serverRegion.AddCommand(serverRegionLiveMigrate)

	// server migrations flags
	serverMigrationList.Flags().String("status", "", "List migration matched by status")
	serverMigrationList.Flags().String("type", "", "List migration matched by type")
	serverMigrationList.Flags().Bool("latest", false, "List latest migrations")
	serverMigrationList.Flags().Bool("watch", false, "List additional fields in output")
	serverMigrationList.Flags().Uint16P("watch-interval", "i", 2, "Loop interval")
	serverMigrationList.Flags().BoolP("long", "l", false, "List loop")

	serverMigration.AddCommand(serverMigrationList)

	// Server prune flags
	serverPrune.Flags().StringP("name", "n", "", "Search by server name")
	serverPrune.Flags().String("host", "", "Search by hostname")
	serverPrune.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverPrune.Flags().BoolP("wait", "w", false, "等待虚拟删除完成")
	serverPrune.Flags().BoolP("yes", "y", false, "所有问题自动回答'是'")

	serverSetPassword.Flags().String("user", "", "User name")

	serverSet.AddCommand(serverSetPassword)

	Server.AddCommand(
		serverList, serverShow, serverCreate, serverDelete, serverPrune,
		serverSet, serverStop, serverStart, serverReboot,
		serverPause, serverUnpause, serverShelve, serverUnshelve,
		serverSuspend, serverResume, serverResize, serverRebuild,
		serverEvacuate, serverMigrate,
		serverMigration,
		serverRegion)
}
