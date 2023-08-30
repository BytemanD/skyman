package identity

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

var Endpoint = &cobra.Command{Use: "endpoint"}

var endpointList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		endpointRegion, _ := cmd.Flags().GetString("region")
		endpointInterface, _ := cmd.Flags().GetString("interface")
		serviceName, _ := cmd.Flags().GetString("service")

		query := url.Values{}
		if endpointRegion != "" {
			query.Set("region_id", endpointRegion)
		}
		if endpointInterface != "" {
			query.Set("interface", endpointInterface)
		}

		client := cli.GetClient()
		serviceMap := map[string]identity.Service{}
		if serviceName != "" {
			services, err := client.Identity.ServiceListByName(serviceName)
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
		services, err := client.Identity.EndpointList(query)
		if err != nil {
			logging.Fatal("get services failed, %s", err)
		}
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "RegionId", Sort: true},
				{Name: "ServiceName", Sort: true, Slot: func(item interface{}) interface{} {
					p, _ := item.(identity.Endpoint)
					if _, ok := serviceMap[p.ServiceId]; !ok {
						service, _ := client.Identity.ServiceShow(p.ServiceId)
						serviceMap[p.ServiceId] = *service
					}
					return serviceMap[p.ServiceId].Name
				}},
				{Name: "Interface"}, {Name: "Url"},
			},
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	endpointList.Flags().BoolP("long", "l", false, "List additional fields in output")
	endpointList.Flags().StringP("region", "r", "", "Search by region Id")
	endpointList.Flags().StringP("interface", "i", "", "Search by interface")
	endpointList.Flags().StringP("service", "s", "", "Search by service name")

	Endpoint.AddCommand(endpointList)
}
