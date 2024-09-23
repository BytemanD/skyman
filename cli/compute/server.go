package compute

import (
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

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/cli/views"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model/glance"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var Server = &cobra.Command{Use: "server"}

func refreshServers(c *openstack.Openstack, filterHosts []string, host, az string, query url.Values) []nova.Server {
	items := []nova.Server{}
	if host != "" || az != "" {
		if len(filterHosts) == 0 {
			logging.Fatal("hosts matched is none")
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

var serverList = &cobra.Command{
	Use:   "list",
	Short: i18n.T("listServers"),
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient()

		query := url.Values{}
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		statusList, _ := cmd.Flags().GetStringArray("status")
		flavor, _ := cmd.Flags().GetString("flavor")
		project, _ := cmd.Flags().GetString("project")
		az, _ := cmd.Flags().GetString("az")
		all, _ := cmd.Flags().GetBool("all")
		dsc, _ := cmd.Flags().GetBool("dsc")
		search, _ := cmd.Flags().GetString("search")
		fields, _ := cmd.Flags().GetString("fields")

		long, _ := cmd.Flags().GetBool("long")
		// verbose, _ := cmd.Flags().GetBool("verbose")
		watch, _ := cmd.Flags().GetBool("watch")
		watchInterval, _ := cmd.Flags().GetUint16("watch-interval")

		if name != "" {
			query.Set("name", name)
		}
		if all {
			query.Set("all_tenants", strconv.FormatBool(all))
		}
		for _, status := range statusList {
			query.Add("status", status)
		}
		if flavor != "" {
			flavor, err := c.NovaV2().Flavor().Found(flavor, false)
			if err != nil {
				logging.Fatal("%s", err)
			}
			query.Set("flavor", flavor.Id)
		}
		filterHosts := utility.Split(host, ",")
		if project != "" {
			p, err := c.KeystoneV3().Project().Found(project)
			utility.LogIfError(err, true, "get project %s failed", project)
			query.Set("tenant_id", p.Id)
			if !all {
				logging.Warning("--all is not set, options --project mybe ignore")
			}
		}
		if az != "" {
			computeService, err := c.NovaV2().Service().ListCompute()
			utility.LogIfError(err, true, "list compute service failed")

			computeService = utility.Filter[nova.Service](
				computeService,
				func(x nova.Service) bool {
					return x.Zone == az
				},
			)
			azHosts := []string{}
			for _, s := range computeService {
				azHosts = append(azHosts, s.Host)
			}
			logging.Debug("hosts matched az %s: %s", az, azHosts)
			if len(azHosts) == 0 {
				logging.Fatal("hosts matched az %s is none", az)
			}
			if len(filterHosts) == 0 {
				filterHosts = azHosts
			} else {
				filterHosts = utility.Filter[string](
					filterHosts,
					func(x string) bool {
						return stringutils.ContainsString(azHosts, x)
					},
				)
			}
		}

		projectMap := map[string]auth.Project{}
		imageMap := map[string]glance.Image{}
		pt := common.PrettyTable{
			Search: search,
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
		if dsc {
			pt.ShortColumns[1].SortMode = table.Dsc
		}
		if long {
			pt.StyleSeparateRows = true
		}
		if fields != "" {
			pt.AddDisplayFields("Id")
			pt.AddDisplayFields(strings.Split(fields, ",")...)
		}
		items := refreshServers(c, filterHosts, host, az, query)
		for {
			pt.AddItems(items)
			common.PrintPrettyTable(pt, long)
			if !watch {
				break
			}
			items = refreshServers(c, filterHosts, host, az, query)
			pt.CleanItems()
			time.Sleep(time.Second * time.Duration(watchInterval))

			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
			var timeNow = time.Now().Format("2006-01-02 15:04:05")
			fmt.Printf("Every %ds\tNow: %s\n", watchInterval, timeNow)
		}
	},
}
var serverShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: i18n.T("showServerDetails"),
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		c := openstack.DefaultClient()
		server, err := c.NovaV2().Server().Found(args[0])
		if err != nil {
			logging.Fatal("%v", err)
		}
		if image, err := c.GlanceV2().Images().Show(server.ImageId()); err == nil {
			server.SetImageName(image.Name)
		}
		views.PrintServer(*server, c)
	},
}
var serverCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "Create server",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")
		if min > max {
			return fmt.Errorf("invalid flags: expect min <= max, got: %d > %d", min, max)
		}
		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetUint16("volume-size")
		if volumeBoot && volumeSize == 0 {
			return fmt.Errorf("invalid flags: --volume-size is required when --volume-boot is true")
		}
		nics, _ := cmd.Flags().GetStringArray("nic")

		if len(nics) > 0 {
			for _, nic := range nics {
				values := strings.Split(nic, "=")
				if len(values) != 2 || values[1] == "" || (values[0] != "net-id" && values[0] != "port-id") {
					return fmt.Errorf("invalid format for flag nic: %s", nic)
				}
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		volumeBoot, _ := cmd.Flags().GetBool("volume-boot")
		volumeSize, _ := cmd.Flags().GetUint16("volume-size")
		volumeType, _ := cmd.Flags().GetString("volume-type")
		flavor, _ := cmd.Flags().GetString("flavor")
		image, _ := cmd.Flags().GetString("image")
		az, _ := cmd.Flags().GetString("az")
		userDataFile, _ := cmd.Flags().GetString("user-data")
		keyName, _ := cmd.Flags().GetString("key-name")
		adminPass, _ := cmd.Flags().GetString("admin-pass")

		min, _ := cmd.Flags().GetUint16("min")
		max, _ := cmd.Flags().GetUint16("max")
		nics, _ := cmd.Flags().GetStringArray("nic")

		wait, _ := cmd.Flags().GetBool("wait")

		createOption := nova.ServerOpt{
			Name:             name,
			Flavor:           flavor,
			AvailabilityZone: az,
			MinCount:         min,
			MaxCount:         max,
		}

		client := openstack.DefaultClient()
		if userDataFile != "" {
			content, err := utility.LoadUserData(userDataFile)
			utility.LogError(err, "read user data failed", true)
			createOption.UserData = content
		}
		if keyName != "" {
			createOption.KeyName = keyName
		}
		if adminPass != "" {
			createOption.AdminPass = adminPass
		}

		if volumeBoot {
			createOption.BlockDeviceMappingV2 = []nova.BlockDeviceMappingV2{
				{
					BootIndex:  0,
					UUID:       image,
					VolumeSize: volumeSize,
					SourceType: "image", DestinationType: "volume",
					DeleteOnTemination: true,
				},
			}
			if volumeType != "" {
				createOption.BlockDeviceMappingV2[0].VolumeType = volumeType
			}
		} else {
			createOption.Image = image
		}
		if len(nics) > 0 {
			createOption.Networks = nova.ParseServerOptNetworks(nics)
		}
		logging.Debug("networks %v", createOption.Networks)
		server, err := client.NovaV2().Server().Create(createOption)
		utility.LogError(err, "create server failed", true)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		server, err = client.NovaV2().Server().Show(server.Id)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		views.PrintServer(*server, nil)
		if wait {
			_, err := client.NovaV2().Server().WaitStatus(server.Id, "ACTIVE", 5)
			if err != nil {
				logging.Error("Server %s create failed, %v", server.Id, err)
			} else {
				logging.Info("Server %s created", server.Id)
			}
		}
	},
}

var serverDelete = &cobra.Command{
	Use:   "delete <server1> [server2 ...]",
	Short: "Delete server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		wait, _ := cmd.Flags().GetBool("wait")
		client := openstack.DefaultClient()

		for _, idOrName := range args {
			client.NovaV2().Server().Found(idOrName)
			s, err := client.NovaV2().Server().Found(idOrName)
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
		if wait {
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
		client := openstack.DefaultClient()
		for _, id := range args {
			err := client.NovaV2().Server().Stop(id)
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
		client := openstack.DefaultClient()
		for _, id := range args {
			err := client.NovaV2().Server().Start(id)
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
		client := openstack.DefaultClient()
		hard, _ := cmd.Flags().GetBool("hard")
		wait, _ := cmd.Flags().GetBool("wait")

		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Found(idOrName)
			utility.LogIfError(err, true, "get server %s failed:", idOrName)

			err = client.NovaV2().Server().Reboot(server.Id, hard)
			if err != nil {
				logging.Fatal("Reqeust to reboot server failed, %v", err)
			} else {
				fmt.Printf("Requested to reboot server: %s\n", server.Id)
			}
		}
		if wait {
			for _, id := range args {
				_, err := client.NovaV2().Server().WaitStatus(id, "ACTIVE", 5)
				if err == nil {
					logging.Info("[%s] rebooted", id)
				} else {
					logging.Error("[%s] reboot failed, %v", id, err)
				}
			}
		}
	},
}
var serverPause = &cobra.Command{
	Use:   "pause <server> [<server> ...]",
	Short: "Pause server(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Pause(id)
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
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Unpause(id)
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
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, id := range args {
			server, err := client.NovaV2().Server().Found(id)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("get server %s faield", id), false)
				continue
			}
			err = client.NovaV2().Server().Shelve(server.Id)
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
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, idorName := range args {
			server, err := client.NovaV2().Server().Found(idorName)
			utility.LogIfError(err, true, "get server %s faield", idorName)
			err = client.NovaV2().Server().Unshelve(server.Id)
			if err != nil {
				logging.Error("Reqeust to unshelve server failed, %v", err)
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
		client := openstack.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Suspend(id)
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
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		for _, id := range args {
			err := client.NovaV2().Server().Resume(id)
			if err != nil {
				logging.Error("Reqeust to resume server failed, %v", err)
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
		confirm, _ := cmd.Flags().GetBool("confirm")
		revert, _ := cmd.Flags().GetBool("revert")
		flavorId, _ := cmd.Flags().GetString("flavor")

		if confirm && revert {
			return fmt.Errorf("flag --confirm and --revert are confict")
		}
		if (!confirm && !revert) && flavorId == "" {
			return fmt.Errorf("flavor is required")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		wait, _ := cmd.Flags().GetBool("wait")
		confirm, _ := cmd.Flags().GetBool("confirm")
		revert, _ := cmd.Flags().GetBool("revert")
		client := openstack.DefaultClient()
		flavorId, _ := cmd.Flags().GetString("flavor")
		var (
			flavor *nova.Flavor
		)
		servers := []nova.Server{}
		for _, serverId := range args {
			server, err := client.NovaV2().Server().Found(serverId)
			if err != nil {
				utility.LogIfError(err, true, "get server %s failed", serverId)
			}
			if confirm {
				if err := client.NovaV2().Server().ResizeConfirm(server.Id); err != nil {
					utility.LogError(err, "Reqeust to confirm resize for server failed", false)
				} else {
					logging.Info("requested to confirm resize for server %s", serverId)
				}
				continue
			}
			if revert {
				if err := client.NovaV2().Server().ResizeRevert(server.Id); err != nil {
					utility.LogError(err, "Reqeust to revert resize for server failed", false)

				} else {
					logging.Info("requested to revert resize for server %s", serverId)
				}
				continue
			}
			if flavorId != "" && flavor == nil {
				flavor, err = client.NovaV2().Flavor().Show(flavorId)
				utility.LogError(err, fmt.Sprintf("Get flavor %s failed", flavorId), true)
			}
			logging.Info("[%s] flavor is %s", server.Id, server.Flavor.OriginalName)
			err = client.NovaV2().Server().Resize(server.Id, flavor.Id)
			if err != nil {
				utility.LogIfError(err, false, "[%s] reqeust to resize server failed", server.Id)
			} else {
				logging.Info("[%s] requested to resize", serverId)
				servers = append(servers, *server)
			}
		}

		if flavorId != "" && wait {
			for _, server := range servers {
				_, err := client.NovaV2().Server().WaitResized(server.Id, flavor.Name)
				if err != nil {
					utility.LogIfError(err, false, "server %s resize failed: %s", server.Id, err)
				} else {
					logging.Success("[%s] resize success", server.Id)
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
		client := openstack.DefaultClient()
		live, _ := cmd.Flags().GetBool("live")
		host, _ := cmd.Flags().GetString("host")
		blockMigrate, _ := cmd.Flags().GetBool("block-migrate")
		wait, _ := cmd.Flags().GetBool("wait")

		srcHostMap := map[string]string{}
		servers := []*nova.Server{}
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Found(idOrName)
			utility.LogError(err, "get server server failed", true)
			servers = append(servers, server)
			logging.Info("[%s] source host is %s", server.Id, server.Host)
			if live {
				srcHostMap[server.Id] = server.Host
				err = client.NovaV2().Server().LiveMigrate(server.Id, blockMigrate, host)
			} else {
				err = client.NovaV2().Server().Migrate(server.Id, host)
			}
			if err != nil {
				utility.LogError(err, "Reqeust to migrate server failed", false)
			} else {
				logging.Info("[%s] requested to migrate server", server.Id)
			}
		}
		if wait {
			for _, server := range servers {
				server, err := client.NovaV2().Server().WaitTask(server.Id, "")
				if err != nil {
					logging.Error("[%s] migrate failed: %s", server.Id, err)
					continue
				}
				if server.Host == srcHostMap[server.Id] {
					logging.Error("[%s] migrate failed, host not changed", server.Id)
				} else {
					logging.Success("[%s] migrate success, %s -> %s",
						server.Id, srcHostMap[server.Id], server.Host)
				}

			}
		}
	},
}

var serverRebuild = &cobra.Command{
	Use:   "rebuild <server>",
	Short: "Rebuild server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		imageIdOrName, _ := cmd.Flags().GetString("image")
		rebuildPassword, _ := cmd.Flags().GetString("rebuild-password")
		name, _ := cmd.Flags().GetString("name")

		server, err := client.NovaV2().Server().Found(args[0])
		utility.LogIfError(err, true, "get server %s failed", args[0])

		options := map[string]interface{}{}
		if imageIdOrName != "" {
			image, err := client.GlanceV2().Images().Found(imageIdOrName)
			utility.LogError(err, "get image failed", true)
			options["imageRef"] = image.Id
		}
		if rebuildPassword != "" {
			options["adminPass"] = rebuildPassword
		}
		if name != "" {
			options["name"] = name
		}

		err = client.NovaV2().Server().Rebuild(server.Id, options)
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
		client := openstack.DefaultClient()
		password, _ := cmd.Flags().GetString("password")
		host, _ := cmd.Flags().GetString("host")
		force, _ := cmd.Flags().GetBool("force")

		err := client.NovaV2().Server().Evacuate(args[0], password, host, force)
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
		serverStatus, _ := cmd.Flags().GetString("status")
		if serverStatus != "" && serverStatus != "active" && serverStatus != "error" {
			return fmt.Errorf("flag --status must be active or error")
		}
		password, _ := cmd.Flags().GetString("password")
		passwordPrompt, _ := cmd.Flags().GetBool("password-prompt")
		if passwordPrompt && password != "" {
			return fmt.Errorf("flag --password and password-prompt is confict")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		serverName, _ := cmd.Flags().GetString("name")
		passwordPrompt, _ := cmd.Flags().GetBool("password-prompt")
		serverPassword, _ := cmd.Flags().GetString("password")
		user, _ := cmd.Flags().GetString("user")
		serverStatus, _ := cmd.Flags().GetString("status")
		description, _ := cmd.Flags().GetString("description")

		if passwordPrompt {
			serverPassword = getPasswordInput()
		}
		params := map[string]interface{}{}
		if serverName != "" {
			params["name"] = serverName
		}
		if description != "" {
			params["description"] = description
		}

		client := openstack.DefaultClient()
		for _, idOrName := range args {
			server, err := client.NovaV2().Server().Found(idOrName)
			utility.LogError(err, "get server failed", true)

			if len(params) > 0 {
				err = client.NovaV2().Server().Set(server.Id, params)
				utility.LogIfError(err, false, "set server name failed for %s", idOrName)
			}
			if serverStatus != "" {
				if serverStatus == "active" {
					err = client.NovaV2().Server().SetState(server.Id, true)
				} else if server.Status == "error" {
					err = client.NovaV2().Server().SetState(server.Id, false)
				}
				utility.LogIfError(err, false, "set server status failed for %s", idOrName)
			}
			if serverPassword != "" {
				err = client.NovaV2().Server().SetPassword(server.Id, serverPassword, user)
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
	return string(newPasswd)
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
		client := openstack.DefaultClient()
		var (
			migrateResp *nova.RegionMigrateResp
			migrateErr  error
		)
		if live {
			migrateResp, migrateErr = client.NovaV2().Server().RegionLiveMigrate(
				serverId, destRegion, true, dryRun, host)
		}
		if migrateErr != nil {
			logging.Fatal("Reqeust to migrate server failed, %v", migrateErr)
		}
		if dryRun {
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
		client := openstack.DefaultClient()
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
			migrations, err := client.NovaV2().Server().ListMigrations(serverId, query)
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
	serverList.Flags().String("host", "", "Search by hostnam, split with ','")
	serverList.Flags().BoolP("all", "a", false, "Display servers from all tenants")
	serverList.Flags().StringArrayP("status", "s", nil, "Search by server status")
	serverList.Flags().String("flavor", "", "Search by flavor")
	serverList.Flags().String("project", "", "Search by project")
	serverList.Flags().String("az", "", "Search by availability zone (admin)")
	serverList.Flags().BoolP("verbose", "v", false, "List verbose fields in output")
	serverList.Flags().Bool("dsc", false, "Sort name by dsc")
	serverList.Flags().String("search", "", i18n.T("localFuzzySearch"))
	serverList.Flags().Bool("watch", false, "List loop")
	serverList.Flags().Uint16P("watch-interval", "i", 2, "Loop interval")
	serverList.Flags().String("fields", "", "Show specified fields")

	// Server create flags
	serverCreate.Flags().String("flavor", "", "Create server with this flavor")
	serverCreate.Flags().StringP("image", "i", "", "Create server with this image")
	serverCreate.Flags().StringArray("nic", []string{},
		"Create NICs on the server. NIC format:\n"+
			"net-id=<net-uuid>: attach NIC to network with this UUID\n"+
			"port-id=<port-uuid>: attach NIC to port with this UUID")
	serverCreate.Flags().Bool("volume-boot", false, "Boot with volume")
	serverCreate.Flags().Uint16("volume-size", 0, "Volume size(GB)")
	serverCreate.Flags().String("volume-type", "", "Volume type.")
	serverCreate.Flags().String("az", "", "Select an availability zone for the server.")
	serverCreate.Flags().Uint16("min", 1, "Minimum number of servers to launch.")
	serverCreate.Flags().Uint16("max", 1, "Maximum number of servers to launch.")
	serverCreate.Flags().String("user-data", "", "user data file to pass to be exposed by the metadata server.")
	serverCreate.Flags().String("key-name", "", "Keypair to inject into this server.")
	serverCreate.Flags().String("admin-pass", "", "Admin password for the instance.")
	serverCreate.Flags().BoolP("wait", "w", false, "Wait server created")

	serverCreate.MarkFlagRequired("flavor")
	serverCreate.MarkFlagRequired("image")

	// server delete flags
	serverDelete.Flags().BoolP("wait", "w", false, "Wait server rebooted")

	// server reboot flags
	serverReboot.Flags().Bool("hard", false, "Perform a hard reboot")
	serverReboot.Flags().BoolP("wait", "w", false, "Wait server rebooted")

	// server migrate flags
	serverMigrate.Flags().Bool("live", false, "Migrate running server.")
	serverMigrate.Flags().String("host", "", "Destination host name.")
	serverMigrate.Flags().Bool("block-migrate", false, "True in case of block_migration.")
	serverMigrate.Flags().Bool("wait", false, "Wait server migrated")

	// server resize flags
	serverResize.Flags().BoolP("wait", "w", false, "Wait server resize completed")
	serverResize.Flags().String("flavor", "", "Resize server to specified flavor")
	serverResize.Flags().Bool("confirm", false, "Confirm server resize is complete")
	serverResize.Flags().Bool("revert", false, "Restore server state before resize")

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

	serverMigration.AddCommand(serverMigrationList)

	serverSet.Flags().String("name", "", "Server name")
	serverSet.Flags().String("password", "", "User password")
	serverSet.Flags().Bool("password-prompt", false, "User password")
	serverSet.Flags().String("user", "", "Username")
	serverSet.Flags().String("status", "", "Server status, active or error")
	serverSet.Flags().String("description", "", "Server description")

	serverRebuild.Flags().String("image", "", "Name or ID of server.")
	serverRebuild.Flags().String("rebuild-password", "", " Set the provided admin password on the rebuilt server.")
	serverRebuild.Flags().String("name", "", "Name for the new server.")

	common.RegistryLongFlag(serverList, serverMigrationList)

	Server.AddCommand(
		serverList, serverShow, serverCreate, serverDelete,
		serverSet, serverStop, serverStart, serverReboot,
		serverPause, serverUnpause, serverShelve, serverUnshelve,
		serverSuspend, serverResume, serverResize, serverRebuild,
		serverEvacuate, serverMigrate,
		serverMigration,
		serverRegion,
	)
}
