package keystone

import (
	"fmt"
	"net/url"
	"os"

	"github.com/BytemanD/go-console/console"
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

var Endpoint = &cobra.Command{Use: "endpoint"}

var endpointList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(0)(cmd, args); err != nil {
			return err
		}
		current, _ := cmd.Flags().GetBool("current")
		region, _ := cmd.Flags().GetString("region")
		if current && region != "" {
			return fmt.Errorf("flags --current and --region conflict")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		current, _ := cmd.Flags().GetBool("current")
		endpointRegion, _ := cmd.Flags().GetString("region")
		endpointInterface, _ := cmd.Flags().GetString("interface")
		serviceName, _ := cmd.Flags().GetString("service")

		query := url.Values{}
		if current {
			endpointRegion = common.CONF.Auth.Region.Id
		}
		if endpointRegion != "" {
			query.Set("region_id", endpointRegion)
		}
		if endpointInterface != "" {
			query.Set("interface", endpointInterface)
		}

		c := openstack.DefaultClient().KeystoneV3()

		serviceMap := map[string]keystone.Service{}
		if serviceName != "" {
			services, err := c.Service().ListByName(serviceName)
			if err != nil {
				console.Error("get service '%s' failed, %v", serviceName, err)
				os.Exit(1)
			}
			if len(services) == 0 {
				console.Error("service '%s' not found", serviceName)
				os.Exit(1)
			}
			for _, service := range services {
				serviceMap[service.Id] = service
				query.Add("service_id", service.Id)
			}
		}
		items, err := c.Endpoint().List(query)
		if err != nil {
			console.Error("list endpoints failed, %s", err)
			os.Exit(1)
		}

		// TODO: 优化
		serviceIds := []string{}
		for _, endpoint := range items {
			found := false
			for _, id := range serviceIds {
				if id == endpoint.ServiceId {
					found = true
					break
				}
			}
			if found {
				continue
			}
			serviceIds = append(serviceIds, endpoint.ServiceId)
		}
		services, err := c.Service().List(nil)
		utility.LogError(err, "get services failed", true)
		for _, service := range services {
			serviceMap[service.Id] = service
		}

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "RegionId", Sort: true},
				{Name: "Service", Sort: true, Slot: func(item interface{}) interface{} {
					p, _ := item.(keystone.Endpoint)
					if service, ok := serviceMap[p.ServiceId]; ok {
						return service.NameOrId()
					} else {
						return p.ServiceId
					}
				}},
				{Name: "Interface"}, {Name: "Url"},
			},
		}
		pt.AddItems(items)
		common.PrintPrettyTable(pt, long)
	},
}
var endpointCreate = &cobra.Command{
	Use:   "create <service> <url>",
	Short: "create endpoint",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		disable, _ := cmd.Flags().GetBool("disable")

		region, _ := cmd.Flags().GetString("region")
		public, _ := cmd.Flags().GetBool("public")
		internal, _ := cmd.Flags().GetBool("internal")
		admin, _ := cmd.Flags().GetBool("admin")

		s, url := args[0], args[1]

		c := openstack.DefaultClient().KeystoneV3()

		service, err := c.Service().Find(s)
		utility.LogError(err, "get service failed", true)
		interfaceMap := map[string]bool{
			"public":   public,
			"admin":    admin,
			"internal": internal,
		}
		endpoints := []keystone.Endpoint{}
		for k, v := range interfaceMap {
			if !v {
				continue
			}
			e := keystone.Endpoint{
				Region:    region,
				Url:       url,
				ServiceId: service.Id,
				Interface: k,
				Enabled:   !disable,
			}
			console.Info("create %s endpoint", k)
			endpoint, err := c.Endpoint().Create(e)
			if err != nil {
				utility.LogError(err, fmt.Sprintf("create %s endpoint failed", k), false)
				continue
			}
			endpoints = append(endpoints, *endpoint)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "RegionId", Sort: true},
				{Name: "Service", Sort: true, Slot: func(item interface{}) interface{} {
					return service.NameOrId()
				}},
				{Name: "Interface"}, {Name: "Url"},
			},
		}
		pt.AddItems(endpoints)
		common.PrintPrettyTable(pt, false)
	},
}
var endpointDelete = &cobra.Command{
	Use:   "delete <endpoint> [endpoint ...]",
	Short: "Delete endpoint",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().KeystoneV3()
		for _, id := range args {
			console.Info("request to delete endpoint %s", id)
			err := c.Endpoint().Delete(id)
			utility.LogError(err, fmt.Sprintf("delete endpoint %s failed", id), false)
		}
	},
}

func init() {
	endpointList.Flags().BoolP("long", "l", false, "List additional fields in output")
	endpointList.Flags().StringP("region", "r", "", "Search by region Id")
	endpointList.Flags().StringP("interface", "i", "", "Search by interface")
	endpointList.Flags().StringP("service", "s", "", "Search by service name")
	endpointList.Flags().Bool("current", false, "Search by current region")

	endpointCreate.Flags().Bool("diable", false, "Disable service")
	endpointCreate.Flags().StringP("region", "r", "", "New endpoint region ID")
	endpointCreate.Flags().Bool("public", false, "Create public endpoint")
	endpointCreate.Flags().Bool("admin", false, "Create admin endpoint")
	endpointCreate.Flags().Bool("internal", false, "Create internal endpoint")

	endpointCreate.MarkFlagRequired("region")

	Endpoint.AddCommand(endpointList, endpointCreate, endpointDelete)
}
