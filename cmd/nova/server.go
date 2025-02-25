package nova

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/howeyc/gopass"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/wxnacy/wgo/file"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/cmd/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var Server = &cobra.Command{Use: "server"}

func refreshServers(c *openstack.Openstack, filterHosts []string, host, az string, query url.Values) []nova.Server {
	items := []nova.Server{}
	if host != "" || az != "" {
		if len(filterHosts) == 0 {
			console.Fatal("hosts matched is none")
		}
		for _, h := range filterHosts {
			if h == "" {
				continue
			}
			query.Set("host", h)
			tmpItems, err := c.NovaV2().Server().Detail(query)
			utility.LogError(err, "list servers failed", true)
			items = append(items, tmpItems...)
		}
	} else {
		tmpItems, err := c.NovaV2().Server().Detail(query)
		utility.LogError(err, "list servers failed", true)
		items = append(items, tmpItems...)
	}
	return items
}

var (
	listFlags                flags.ServerListFlags
	setFlags                 flags.ServerSetFlags
	deleteFlags              flags.ServerDeleteFlags
	createFlags              flags.ServerCreateFlags
	rebootFlags              flags.ServerRebootFlags
	resizeFlags              flags.ServerResizeFlags
	migrateFlags             flags.ServerMigrateFlags
	rebuildFlags             flags.ServerRebuildFlags
	evacuateFlags            flags.ServerEvacuateFlags
	regionMigrateFlags       flags.ServerRegionMigrateFlags
	serverMigrationListFlags flags.ServerMigrationListFlags
	createImageFlags         flags.ServerCreateImageFlags
)

var serverList = &cobra.Command{
	Use:   "list",
	Short: i18n.T("listServers"),
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient()

		query := url.Values{}
		if *listFlags.Name != "" {
			query.Set("name", *listFlags.Name)
		}
		if *listFlags.All {
			query.Set("all_tenants", "true")
		}
		for _, status := range *listFlags.Status {
			query.Add("status", status)
		}
		if *listFlags.Flavor != "" {
			flavor, err := c.NovaV2().Flavor().Find(*listFlags.Flavor, false)
			if err != nil {
				console.Fatal("%s", err)
			}
			query.Set("flavor", flavor.Id)
		}
		filterHosts := utility.Split(*listFlags.Host, ",")
		if *listFlags.Project != "" {
			p, err := c.KeystoneV3().Project().Find(*listFlags.Project)
			utility.LogIfError(err, true, "get project %s failed", *listFlags.Project)
			query.Set("tenant_id", p.Id)
			if !*listFlags.All {
				console.Fatal("--all is not set, options --project mybe ignore")
			}
		}
		if *listFlags.AZ != "" {
			computeService, err := c.NovaV2().Service().List(nil)
			utility.LogIfError(err, true, "list compute service failed")

			computeService = lo.Filter(computeService, func(x nova.Service, _ int) bool {
				return x.Zone == *listFlags.AZ
			})
			azHosts := lo.Map(computeService, func(s nova.Service, _ int) string {
				return s.Host
			})
			console.Debug("hosts matched az %s: %s", *listFlags.AZ, azHosts)
			if len(azHosts) == 0 {
				console.Fatal("hosts matched az %s is none", *listFlags.AZ)
			}
			if len(filterHosts) == 0 {
				filterHosts = azHosts
			} else {
				filterHosts = lo.Filter(filterHosts, func(x string, _ int) bool {
					return lo.Contains(azHosts, x)
				})
			}
		}

		projectMap := map[string]model.Project{}
		imageMap := map[string]glance.Image{}
		pt := common.PrettyTable{
			Search: *listFlags.Search,
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true}, {Name: "Status", AutoColor: true},
				{Name: "TaskState"},
				{Name: "PowerState", AutoColor: true, Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					return p.GetPowerState()
				}},
				{Name: "Addresses", Text: "Networks", WidthMax: 70, Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					return strings.Join(p.GetNetworks(), "\n")
				}},
				{Name: "Host"},
			},
			LongColumns: []common.Column{
				{Name: "AZ", Text: "AZ"}, {Name: "InstanceName"},
				{Name: "Flavor", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					return p.Flavor.OriginalName
				}},
				{Name: "Project", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					if project, ok := projectMap[p.TenantId]; ok {
						return project.Name
					}
					if project, err := c.KeystoneV3().Project().Show(p.TenantId); err == nil {
						projectMap[p.TenantId] = *project
						return project.Name
					}
					return p.TenantId
				}},
				{Name: "Vcpus", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					return p.Flavor.Vcpus
				}},
				{Name: "Ram", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					return p.Flavor.Ram
				}},
				{Name: "Image", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Server)
					imageId := p.ImageId()
					if imageId == "" {
						return ""
					}
					if image, ok := imageMap[imageId]; ok {
						return image.Name
					}
					if image, err := c.GlanceV2().Images().Show(imageId); err == nil {
						imageMap[imageId] = *image
						return image.Name
					}
					return imageId
				}},
			},
		}
		if *listFlags.Dsc {
			pt.ShortColumns[1].SortMode = table.Dsc
		}
		if *listFlags.Long {
			pt.StyleSeparateRows = true
		}
		if *listFlags.Fields != "" {
			pt.AddDisplayFields("Id")
			pt.AddDisplayFields(strings.Split(*listFlags.Fields, ",")...)
		}
		items := refreshServers(c, filterHosts, *listFlags.Host, *listFlags.AZ, query)
		for {
			pt.AddItems(items)
			common.PrintPrettyTable(pt, *listFlags.Long)
			if !*listFlags.Watch {
				break
			}
			items = refreshServers(c, filterHosts, *listFlags.Host, *listFlags.AZ, query)
			pt.CleanItems()
			time.Sleep(time.Second * time.Duration(*listFlags.WatchInterval))

			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
			var timeNow = time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("Every %ds\tNow: %s\n", *listFlags.WatchInterval, timeNow)
		}
	},
}
var serverShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: i18n.T("showServerDetails"),
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		c := common.DefaultClient()
		server, err := c.NovaV2().Server().Find(args[0])
		if err != nil {
			console.Fatal("%v", err)
		}
		if image, err := c.GlanceV2().Images().Show(server.ImageId()); err == nil {
			server.SetImageName(image.Name)
		}
		views.PrintServer(*server, c)
	},
}

const nicUsage = `
Create NICs on the server. NIC format:
  net-id=<net-uuid>: attach NIC to network with this UUID
  port-id=<port-uuid>: attach NIC to port with this UUID
`
const createExample = `
server create demo --flavor 1g1v --image cirros
server create demo --flavor 1g1v --image cirros --volume-boot
server create demo --flavor 1g1v --image cirros --volume-boot --nic net-id=<network id>
`

var serverCreate = &cobra.Command{
	Use:     "create <name>",
	Short:   "Create server",
	Example: strings.Trim(createExample, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}

		if *createFlags.Max > *createFlags.Min {
			return fmt.Errorf("invalid flags: expect min <= max, got: %d > %d", *createFlags.Max, *createFlags.Min)
		}
		if *createFlags.VolumeBoot && *createFlags.VolumeSize == 0 {
			return fmt.Errorf("invalid flags: --volume-size is required when --volume-boot is true")
		}

		for _, nic := range *createFlags.Nic {
			values := strings.Split(nic, "=")
			if len(values) != 2 || values[1] == "" || (values[0] != "net-id" && values[0] != "port-id") {
				return fmt.Errorf("invalid format for flag nic: %s", nic)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		createOption := nova.ServerOpt{
			Name:             args[0],
			Flavor:           *createFlags.Flavor,
			AvailabilityZone: *createFlags.AZ,
			MinCount:         *createFlags.Min,
			MaxCount:         *createFlags.Max,
		}

		client := common.DefaultClient()
		if *createFlags.UserData != "" {
			content, err := utility.LoadUserData(*createFlags.UserData)
			utility.LogError(err, "read user data failed", true)
			createOption.UserData = content
		}
		if *createFlags.KeyName != "" {
			createOption.KeyName = *createFlags.KeyName
		}
		if *createFlags.AdminPass != "" {
			createOption.AdminPass = *createFlags.AdminPass
		}

		if *createFlags.VolumeBoot {
			createOption.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
				{
					BootIndex:          0,
					UUID:               *createFlags.Image,
					VolumeSize:         *createFlags.VolumeSize,
					SourceType:         "image",
					DestinationType:    "volume",
					DeleteOnTemination: true,
				},
			}
			if *createFlags.VolumeType != "" {
				createOption.BlockDeviceMappingV2[0].VolumeType = *createFlags.VolumeType
			}
		} else {
			createOption.Image = *createFlags.Image
		}
		if len(*createFlags.Nic) > 0 {
			createOption.Networks = nova.ParseServerOptNetworks(*createFlags.Nic)
		}
		console.Debug("networks %v", createOption.Networks)
		server, err := client.NovaV2().Server().Create(createOption)
		utility.LogError(err, "create server failed", true)
		if err != nil {
			println(err)
			os.Exit(1)
		}
		server, err = client.NovaV2().Server().Show(server.Id)
		if err != nil {
			println(err)
			os.Exit(1)
		}
		views.PrintServer(*server, nil)
		if *createFlags.Wait {
			_, err := client.NovaV2().Server().WaitStatus(server.Id, "ACTIVE", 5)
			if err != nil {
				console.Error("Server %s create failed, %v", server.Id, err)
			} else {
				console.Info("Server %s created", server.Id)
			}
		}
	},
}

var serverDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, idOrName := range args {
			client.NovaV2().Server().Find(idOrName)
			s, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("found server %s failed", idOrName), false)
				continue
			}
			err = client.NovaV2().Server().Delete(s.Id)
			if err != nil {
				utility.LogError(err, "delete server %s failed %s", false)
				continue
			}
			fmt.Printf("Requested to delete server: %s\n", idOrName)
		}
		if *deleteFlags.Wait {
			for _, id := range args {
				client.NovaV2().Server().WaitDeleted(id)
			}
		}
	},
}

var serverStop = &cobra.Command{
	Use:   "stop <server> [<server> ...]",
	Short: "Stop server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				console.Error("get server %s failed, %v", idOrName, err)
				continue
			}
			err = client.NovaV2().Server().Stop(server.Id)
			if err != nil {
				console.Error("Reqeust to stop server %s failed, %v", idOrName, err)
			} else {
				console.Info("Requested to stop server: %s\n", idOrName)
			}
		}
	},
}
var serverStart = &cobra.Command{
	Use:   "start <server> [<server> ...]",
	Short: "Start server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				console.Error("get server %s failed, %v", idOrName, err)
				continue
			}
			err = client.NovaV2().Server().Start(server.Id)
			if err != nil {
				console.Error("Reqeust to start server failed, %v", err)
			} else {
				fmt.Printf("Requested to start server: %s\n", server.Id)
			}
		}
	},
}

var serverReboot = &cobra.Command{
	Use:   "reboot <server> [<server> ...]",
	Short: "Reboot server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		servers := []nova.Server{}
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				console.Error("get server %s failed, %v", idOrName, err)
				continue
			}
			servers = append(servers, *server)
			err = client.NovaV2().Server().Reboot(server.Id, *rebootFlags.Hard)
			if err != nil {
				console.Error("Reqeust to reboot server failed, %v", err)
			} else {
				fmt.Printf("Requested to reboot server: %s\n", server.Id)
			}
		}
		if !*rebootFlags.Wait {
			return
		}
		for _, server := range servers {
			_, err := client.NovaV2().Server().WaitStatus(server.Id, "ACTIVE", 5)
			if err == nil {
				console.Info("[%s] rebooted", server.Id)
			} else {
				console.Error("[%s] reboot failed, %v", server.Id, err)
			}
		}
	},
}
var serverPause = &cobra.Command{
	Use:   "pause <server> [<server> ...]",
	Short: "Pause server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				console.Error("get server %s failed, %v", idOrName, err)
				continue
			}
			err = client.NovaV2().Server().Pause(server.Id)
			if err != nil {
				console.Error("Reqeust to pause server failed, %v", err)
			} else {
				fmt.Printf("Requested to pause server: %s\n", idOrName)
			}
		}
	},
}
var serverUnpause = &cobra.Command{
	Use:   "unpause <server> [<server> ...]",
	Short: "Unpause server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			if err != nil {
				console.Error("get server %s failed, %v", idOrName, err)
				continue
			}
			err = client.NovaV2().Server().Unpause(server.Id)
			if err != nil {
				console.Error("Reqeust to unpause server failed, %v", err)
			} else {
				fmt.Printf("Requested to unpause server: %s\n", server.Id)
			}
		}
	},
}
var serverShelve = &cobra.Command{
	Use:   "shelve <server> [<server> ...]",
	Short: "Shelve server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, id := range args {
			server, err := client.NovaV2().Server().Find(id, client.NovaV2().IsAdmin)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("get server %s faield", id), false)
				continue
			}
			err = client.NovaV2().Server().Shelve(server.Id)
			if err != nil {
				console.Error("Reqeust to shelve server failed, %v", err)
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
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, idorName := range args {
			server, err := client.NovaV2().Server().Find(idorName)
			utility.LogIfError(err, true, "get server %s faield", idorName)
			err = client.NovaV2().Server().Unshelve(server.Id)
			if err != nil {
				console.Error("Reqeust to unshelve server failed, %v", err)
			} else {
				fmt.Printf("Requested to unshelve server: %s\n", idorName)
			}
		}
	},
}
var serverSuspend = &cobra.Command{
	Use:   "suspend <server> [<server> ...]",
	Short: "Suspend server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Suspend(id)
			if err != nil {
				console.Error("Reqeust to susppend server failed, %v", err)
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
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Resume(id)
			if err != nil {
				console.Error("Reqeust to resume server failed, %v", err)
			} else {
				fmt.Printf("Requested to resume server: %s\n", id)
			}
		}
	},
}

var serverResize = &cobra.Command{
	Use:   "resize <server1> [server2 ...]",
	Short: "Resize server",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		if *resizeFlags.Confirm && *resizeFlags.Revert {
			return fmt.Errorf("flag --confirm and --revert are confict")
		}
		if (!*resizeFlags.Confirm && !*resizeFlags.Revert) && *resizeFlags.Flavor == "" {
			return fmt.Errorf("flavor is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()
		servers := []nova.Server{}
		var flavor *nova.Flavor

		for _, serverId := range args {
			server, err := client.NovaV2().Server().Find(serverId)
			if err != nil {
				utility.LogIfError(err, true, "get server %s failed", serverId)
			}
			if *resizeFlags.Confirm {
				if err := client.NovaV2().Server().ResizeConfirm(server.Id); err != nil {
					utility.LogError(err, "Reqeust to confirm resize for server failed", false)
				} else {
					console.Info("requested to confirm resize for server %s", serverId)
				}
				continue
			}
			if *resizeFlags.Revert {
				if err := client.NovaV2().Server().ResizeRevert(server.Id); err != nil {
					utility.LogError(err, "Reqeust to revert resize for server failed", false)

				} else {
					console.Info("requested to revert resize for server %s", serverId)
				}
				continue
			}
			if *resizeFlags.Flavor != "" {
				flavor, err = client.NovaV2().Flavor().Find(*resizeFlags.Flavor, false)
				utility.LogError(err, fmt.Sprintf("Get flavor %s failed", *resizeFlags.Flavor), true)
			}
			console.Info("[%s] flavor is %s", server.Id, server.Flavor.OriginalName)
			err = client.NovaV2().Server().Resize(server.Id, flavor.Id)
			if err != nil {
				utility.LogIfError(err, false, "[%s] reqeust to resize server failed", server.Id)
			} else {
				console.Info("[%s] requested to resize", serverId)
				servers = append(servers, *server)
			}
		}

		if flavor != nil && *resizeFlags.Wait {
			for _, server := range servers {
				_, err := client.NovaV2().Server().WaitResized(server.Id, flavor.Name)
				if err != nil {
					utility.LogIfError(err, false, "server %s resize failed: %s", server.Id, err)
				} else {
					console.Success("[%s] resize success", server.Id)
				}

			}
		}
	},
}
var serverMigrate = &cobra.Command{
	Use:   "migrate <server1> [server2]",
	Short: "Migrate server",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()
		// live, _ := cmd.Flags().GetBool("live")
		// host, _ := cmd.Flags().GetString("host")
		// blockMigrate, _ := cmd.Flags().GetBool("block-migrate")
		// wait, _ := cmd.Flags().GetBool("wait")

		srcHostMap := map[string]string{}
		servers := []*nova.Server{}
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			utility.LogError(err, "get server server failed", true)
			servers = append(servers, server)
			console.Info("[%s] source host is %s", server.Id, server.Host)
			if *migrateFlags.Live {
				srcHostMap[server.Id] = server.Host
				err = client.NovaV2().Server().LiveMigrate(server.Id, *migrateFlags.BlockMigrate, *migrateFlags.Host)
			} else {
				err = client.NovaV2().Server().Migrate(server.Id, *migrateFlags.Host)
			}
			if err != nil {
				utility.LogError(err, "Reqeust to migrate server failed", false)
			} else {
				console.Info("[%s] requested to migrate server", server.Id)
			}
		}
		if *migrateFlags.Wait {
			for _, server := range servers {
				server, err := client.NovaV2().Server().WaitTask(server.Id, "")
				if err != nil {
					console.Error("[%s] migrate failed: %s", server.Id, err)
					continue
				}
				if server.Host == srcHostMap[server.Id] {
					console.Error("[%s] migrate failed, host not changed", server.Id)
				} else {
					console.Success("[%s] migrate success, %s -> %s",
						server.Id, srcHostMap[server.Id], server.Host)
				}

			}
		}
	},
}

var serverRebuild = &cobra.Command{
	Use:   "rebuild <server>",
	Short: "Rebuild server",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		if *rebuildFlags.UserData != "" && *rebuildFlags.UserDataUnset {
			return errors.New("cannot specify '--user-data-unset' with '--user-data'")
		}
		if *rebuildFlags.UserData != "" {
			if !file.IsFile(*rebuildFlags.UserData) {
				return fmt.Errorf("file %s not exists", *rebuildFlags.UserData)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		server, err := client.NovaV2().Server().Find(args[0])
		utility.LogIfError(err, true, "get server %s failed", args[0])

		opt := nova.RebuilOpt{
			Password: *rebuildFlags.Password,
			Name:     *rebuildFlags.Name,
		}
		if *rebuildFlags.Image != "" {
			image, err := client.GlanceV2().Images().Find(*rebuildFlags.Image)
			utility.LogError(err, "get image failed", true)
			opt.ImageId = image.Id
		}

		if *rebuildFlags.UserDataUnset {
			opt.UserData = nil
		} else if *rebuildFlags.UserData != "" {
			content, err := utility.LoadUserData(*rebuildFlags.UserData)
			utility.LogError(err, "read user data failed", true)
			opt.UserData = content
		} else {
			opt.UserData = ""
		}
		err = client.NovaV2().Server().Rebuild(server.Id, opt)
		if err != nil {
			console.Error("Reqeust to rebuild server failed, %v", err)
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
		client := common.DefaultClient()

		err := client.NovaV2().Server().Evacuate(args[0],
			*evacuateFlags.Password, *evacuateFlags.Host, *evacuateFlags.Force)
		if err != nil {
			utility.LogError(err, "Reqeust to evacuate server failed", true)
		} else {
			fmt.Printf("Requested to evacuate server: %s\n", args[0])
		}
	},
}

var serverSet = &cobra.Command{
	Use:   "set <server> [<server> ...]",
	Short: "Update server",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
			return err
		}
		serverStatus := *setFlags.Status
		if serverStatus != "" && serverStatus != "active" && serverStatus != "error" {
			return fmt.Errorf("flag --status must be active or error")
		}
		if *setFlags.PasswordPrompt && *setFlags.Password != "" {
			return fmt.Errorf("flag --password and password-prompt is confict")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		serverPassword := *setFlags.Password

		if *setFlags.PasswordPrompt {
			serverPassword = getPasswordInput()
		}

		params := map[string]interface{}{}
		if *setFlags.Name != "" {
			params["name"] = *setFlags.Name
		}
		if *setFlags.Description != "" {
			params["description"] = *setFlags.Description
		}

		client := common.DefaultClient()
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Find(idOrName)
			utility.LogError(err, "get server failed", true)

			if len(params) > 0 {
				err = client.NovaV2().Server().Set(server.Id, params)
				utility.LogIfError(err, false, "set server name failed for %s", idOrName)
			}

			if *setFlags.Status != "" {
				switch *setFlags.Status {
				case "active":
					err = client.NovaV2().Server().SetState(server.Id, true)
				case "error":
					err = client.NovaV2().Server().SetState(server.Id, false)
				}
				utility.LogIfError(err, false, "set server status failed for %s", idOrName)
			}
			if serverPassword != "" {
				err = client.NovaV2().Server().SetPassword(server.Id, serverPassword, *setFlags.Status)
				utility.LogIfError(err, false, "set server password failed for %s", idOrName)
			}
		}
	},
}

func getPasswordInput() string {
	var newPasswd, again []byte
	for {
		fmt.Printf("New password: ")
		newPasswd, _ = gopass.GetPasswd()
		if string(newPasswd) == "" {
			console.Error("Password is empty.")
			continue
		}
		fmt.Printf("Again: ")
		again, _ = gopass.GetPasswd()
		if string(again) == string(newPasswd) {
			break
		}
		console.Error("Passwords do not match.")
	}
	return string(newPasswd)
}

var serverRegion = &cobra.Command{Use: "region"}
var serverRegionLiveMigrate = &cobra.Command{
	Use:   "migrate <server> <dest region>",
	Short: "Migrate server to another region",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		serverId, destRegion := args[0], args[1]

		if !*regionMigrateFlags.Live {
			console.Error("Only support live migrate now, please use option --live")
		}
		client := common.DefaultClient()
		var (
			migrateResp *nova.RegionMigrateResp
			migrateErr  error
		)
		if *regionMigrateFlags.Live {
			migrateResp, migrateErr = client.NovaV2().Server().RegionLiveMigrate(
				serverId, destRegion, true, *regionMigrateFlags.DryRun, *regionMigrateFlags.Host)
		}
		if migrateErr != nil {
			console.Error("Reqeust to migrate server failed, %v", migrateErr)
		}
		if *regionMigrateFlags.DryRun {
			table := common.PrettyItemTable{
				Item: migrateResp,
				ShortFields: []common.Column{
					{Name: "AllowLiveMigrate"}, {Name: "Reason"},
				},
			}
			common.PrintPrettyItemTable(table)
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
		query := url.Values{}
		if *migrationListFlags.Status != "" {
			query.Set("status", *migrationListFlags.Status)
		}
		if *serverMigrationListFlags.Latest {
			query.Set("latest", "true")
		}
		if *migrationListFlags.Type != "" {
			query.Set("type", *migrationListFlags.Type)
		}
		client := common.DefaultClient()
		server, err := client.NovaV2().Server().Find(args[0])
		if err != nil {
			console.Error("get server %s failed: %s", args[0], err)
		}
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
			migrations, err := client.NovaV2().Server().ListMigrations(server.Id, query)
			if err != nil {
				console.Error("Reqeust to list server migration failed, %v", err)
			}
			table.CleanItems()
			table.AddItems(migrations)
			if *serverMigrationListFlags.Watch {
				cmd := exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
				var timeNow = time.Now().Format("2006-01-02 15:04:05")
				fmt.Printf("Every %ds\tNow: %s\n", *serverMigrationListFlags.WatchInterval, timeNow)
			}
			common.PrintPrettyTable(table, *migrationListFlags.Long)
			if !*serverMigrationListFlags.Watch {
				break
			}
			time.Sleep(time.Second * time.Duration(*serverMigrationListFlags.WatchInterval))
		}
	},
}
var serverImageCmd = &cobra.Command{Use: "image"}
var createImageCmd = &cobra.Command{
	Use:   "create <server> <image name>",
	Short: "Create a new image by taking of a running server.",
	Example: "server image create SERVER image-for-server\n" +
		"server image create SERVER image-for-server --metadata image_type=instance_image",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		for _, metadata := range *createImageFlags.Metadata {
			if !strings.Contains(metadata, "=") {
				return fmt.Errorf("invalid metadata %s, must be key=value", metadata)
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		idOrName, imageName := args[0], args[1]
		client := common.DefaultClient()

		server, err := client.NovaV2().Server().Find(idOrName)
		utility.LogError(err, "get server failed", true)
		metadata := map[string]string{}
		for _, meta := range *createImageFlags.Metadata {
			values := strings.Split(meta, "=")
			metadata[values[0]] = values[1]
		}
		imageId, err := client.NovaV2().Server().CreateImage(server.Id, imageName, metadata)
		utility.LogError(err, "create image failed", true)
		console.Info("requested to create image success, image id: %s", imageId)
	},
}

func init() {
	listFlags = flags.ServerListFlags{
		Name:          serverList.Flags().StringP("name", "n", "", "Search by server name"),
		Host:          serverList.Flags().String("host", "", "Search by hostnam, split with ','"),
		Status:        serverList.Flags().StringArrayP("status", "s", nil, "Search by server status"),
		Flavor:        serverList.Flags().String("flavor", "", "Search by flavor"),
		Project:       serverList.Flags().String("project", "", "Search by project"),
		AZ:            serverList.Flags().String("az", "", "Search by availability zone (admin)"),
		All:           serverList.Flags().BoolP("all", "a", false, "Display servers from all tenants"),
		Verbose:       serverList.Flags().BoolP("verbose", "v", false, "List verbose fields in output"),
		Dsc:           serverList.Flags().Bool("dsc", false, "Sort name by dsc"),
		Search:        serverList.Flags().String("search", "", i18n.T("localFuzzySearch")),
		Watch:         serverList.Flags().Bool("watch", false, "List loop"),
		WatchInterval: serverList.Flags().UintP("watch-interval", "i", 2, "Loop interval"),
		Fields:        serverList.Flags().String("fields", "", "Show specified fields"),
		Long:          serverList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}
	createFlags = flags.ServerCreateFlags{
		Flavor:     serverCreate.Flags().String("flavor", "", "Create server with this flavor"),
		Image:      serverCreate.Flags().StringP("image", "i", "", "Create server with this image"),
		Nic:        serverCreate.Flags().StringArray("nic", []string{}, strings.Trim(nicUsage, "\n")),
		VolumeBoot: serverCreate.Flags().Bool("volume-boot", false, "Boot with volume"),
		VolumeType: serverCreate.Flags().String("volume-type", "", "Volume type."),
		VolumeSize: serverCreate.Flags().Uint16("volume-size", 0, "Volume size(GB)"),
		AZ:         serverCreate.Flags().String("az", "", "Select an availability zone for the server."),
		Min:        serverCreate.Flags().Uint16("min", 1, "Minimum number of servers to launch."),
		Max:        serverCreate.Flags().Uint16("max", 1, "Maximum number of servers to launch."),
		UserData:   serverCreate.Flags().String("user-data", "", "user data file to pass to be exposed by the metadata server."),
		KeyName:    serverCreate.Flags().String("key-name", "", "Keypair to inject into this server."),
		AdminPass:  serverCreate.Flags().String("admin-pass", "", "Admin password for the instance."),
		Wait:       serverCreate.Flags().BoolP("wait", "w", false, "Wait server created"),
	}

	serverCreate.MarkFlagRequired("flavor")
	serverCreate.MarkFlagRequired("image")

	deleteFlags = flags.ServerDeleteFlags{
		Wait: serverDelete.Flags().BoolP("wait", "w", false, "Wait server rebooted"),
	}
	rebootFlags = flags.ServerRebootFlags{
		Hard: serverReboot.Flags().Bool("hard", false, "Perform a hard reboot"),
		Wait: serverReboot.Flags().BoolP("wait", "w", false, "Wait server rebooted"),
	}

	migrateFlags = flags.ServerMigrateFlags{
		Live:         serverMigrate.Flags().Bool("live", false, "Migrate running server."),
		Host:         serverMigrate.Flags().String("host", "", "Destination host name."),
		BlockMigrate: serverMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration."),
		Wait:         serverMigrate.Flags().Bool("wait", false, "Wait server migrated"),
	}

	resizeFlags = flags.ServerResizeFlags{
		Flavor:  serverResize.Flags().String("flavor", "", "Resize server to specified flavor"),
		Confirm: serverResize.Flags().Bool("confirm", false, "Confirm server resize is complete"),
		Revert:  serverResize.Flags().Bool("revert", false, "Restore server state before resize"),
		Wait:    serverResize.Flags().BoolP("wait", "w", false, "Wait server resize completed"),
	}

	evacuateFlags = flags.ServerEvacuateFlags{
		Force:    serverEvacuate.Flags().Bool("force", false, "Force to not verify the scheduler if a host is provided."),
		Host:     serverEvacuate.Flags().String("host", "", "Destination host name."),
		Password: serverEvacuate.Flags().String("password", "", "Set the provided admin password on the evacuated server."),
	}

	regionMigrateFlags = flags.ServerRegionMigrateFlags{
		Live: serverRegionLiveMigrate.Flags().Bool("live", false, "Migrate running server."),
		Host: serverRegionLiveMigrate.Flags().String("host", "", "Destination host name."),
		// serverRegionLiveMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration.")
		DryRun: serverRegionLiveMigrate.Flags().Bool("dry-run", false, "True in case of dry run."),
	}
	serverRegion.AddCommand(serverRegionLiveMigrate)

	serverMigrationListFlags = flags.ServerMigrationListFlags{
		Status:        serverMigrationList.Flags().String("status", "", "List migration matched by status"),
		Type:          serverMigrationList.Flags().String("type", "", "List migration matched by type"),
		Latest:        serverMigrationList.Flags().Bool("latest", false, "List latest migrations"),
		Watch:         serverMigrationList.Flags().Bool("watch", false, "List additional fields in output"),
		WatchInterval: serverMigrationList.Flags().Uint16P("watch-interval", "i", 2, "Loop interval"),
		Long:          serverMigrationList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}

	serverMigration.AddCommand(serverMigrationList)

	setFlags = flags.ServerSetFlags{
		Name:           serverSet.Flags().String("name", "", "Server name"),
		Password:       serverSet.Flags().String("password", "", "User password"),
		PasswordPrompt: serverSet.Flags().Bool("password-prompt", false, "User password"),
		User:           serverSet.Flags().String("user", "", "Username"),
		Status:         serverSet.Flags().String("status", "", "Server status, active or error"),
		Description:    serverSet.Flags().String("description", "", "Server description"),
	}
	rebuildFlags = flags.ServerRebuildFlags{
		Image:         serverRebuild.Flags().String("image", "", "Name or ID of server."),
		Password:      serverRebuild.Flags().String("rebuild-password", "", " Set the provided admin password on the rebuilt server."),
		Name:          serverRebuild.Flags().String("name", "", "Name for the new server."),
		UserData:      serverRebuild.Flags().String("user-data", "", "user data file to pass to be exposed by the metadata server."),
		UserDataUnset: serverRebuild.Flags().Bool("user-data-unset", false, "Unset user_data in the server."),
	}
	createImageFlags = flags.ServerCreateImageFlags{
		Metadata: createImageCmd.Flags().StringArray("metadata", []string{}, "Image metadata, format: key=value"),
	}

	serverImageCmd.AddCommand(createImageCmd)

	Server.AddCommand(
		serverList, serverShow, serverCreate, serverDelete,
		serverSet, serverStop, serverStart, serverReboot,
		serverPause, serverUnpause, serverShelve, serverUnshelve,
		serverSuspend, serverResume, serverResize, serverRebuild,
		serverEvacuate, serverMigrate,
		serverMigration,
		serverRegion,
		serverImageCmd,
	)
}
