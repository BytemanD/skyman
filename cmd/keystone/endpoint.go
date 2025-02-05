package keystone

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/go-console/console"
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
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

		c := common.DefaultClient().KeystoneV3()

		serviceMap := map[string]keystone.Service{}
		if serviceName != "" {
			service, err := c.Service().Find(serviceName)
			utility.LogIfError(err, true, "get service '%s' failed", serviceName)
			serviceMap[service.Id] = *service
			query.Add("service_id", service.Id)
		}
		items, err := c.Endpoint().List(query)
		utility.LogIfError(err, true, "list endpoints failed")
		for _, endpoint := range items {
			if _, ok := serviceMap[endpoint.ServiceId]; ok {
				continue
			}
			services, err := c.Service().List(nil)
			utility.LogError(err, "get services failed", true)
			for _, srv := range services {
				if _, ok := serviceMap[srv.Id]; ok {
					continue
				}
				serviceMap[srv.Id] = srv
			}
		}
		common.PrintEndpoints(items, long, serviceMap)
	},
}
var endpointCreate = &cobra.Command{
	Use:   "create <service> <url>",
	Short: "create endpoint",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(2)(cmd, args); err != nil {
			return err
		}
		public, _ := cmd.Flags().GetBool("public")
		internal, _ := cmd.Flags().GetBool("internal")
		admin, _ := cmd.Flags().GetBool("admin")
		if !public && !internal && !admin {
			return fmt.Errorf("at least one of --public, --internal and --admin needs to be specified")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		disable, _ := cmd.Flags().GetBool("disable")

		region, _ := cmd.Flags().GetString("region")
		public, _ := cmd.Flags().GetBool("public")
		internal, _ := cmd.Flags().GetBool("internal")
		admin, _ := cmd.Flags().GetBool("admin")

		s, url := args[0], args[1]
		c := common.DefaultClient().KeystoneV3()

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
		common.PrintEndpoints(endpoints, false, nil)
	},
}
var endpointDelete = &cobra.Command{
	Use:   "delete <endpoint> [endpoint ...]",
	Short: "Delete endpoint",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().KeystoneV3()
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
