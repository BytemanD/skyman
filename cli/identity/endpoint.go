package identity

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/keystone"
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
			services, err := c.Services().ListByName(serviceName)
			if err != nil {
				logging.Fatal("get service '%s' failed, %v", serviceName, err)
			}
			if len(services) == 0 {
				logging.Fatal("service '%s' not found", serviceName)
			}
			for _, service := range services {
				serviceMap[service.Id] = service
				query.Add("service_id", service.Id)
			}
		}
		items, err := c.Endpoints().List(query)
		if err != nil {
			logging.Fatal("get services failed, %s", err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "RegionId", Sort: true},
				{Name: "ServiceName", Sort: true, Slot: func(item interface{}) interface{} {
					p, _ := item.(keystone.Endpoint)
					if _, ok := serviceMap[p.ServiceId]; !ok {
						service, _ := c.Services().Show(p.ServiceId)
						serviceMap[p.ServiceId] = *service
					}
					return serviceMap[p.ServiceId].Name
				}},
				{Name: "Interface"}, {Name: "Url"},
			},
		}
		pt.AddItems(items)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	endpointList.Flags().BoolP("long", "l", false, "List additional fields in output")
	endpointList.Flags().StringP("region", "r", "", "Search by region Id")
	endpointList.Flags().StringP("interface", "i", "", "Search by interface")
	endpointList.Flags().StringP("service", "s", "", "Search by service name")
	endpointList.Flags().Bool("current", false, "Search by current region")

	Endpoint.AddCommand(endpointList)
}
