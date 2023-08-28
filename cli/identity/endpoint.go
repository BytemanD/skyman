package identity

import (
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
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
		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Id", "RegionId", "ServiceName", "Interface", "Url"},
			SortBy:       []table.SortBy{{Name: "Region"}, {Name: "Service Name"}},
			Slots: map[string]func(item interface{}) interface{}{
				"ServiceName": func(item interface{}) interface{} {
					p, _ := item.(identity.Endpoint)
					if _, ok := serviceMap[p.ServiceId]; !ok {
						service, _ := client.Identity.ServiceShow(p.ServiceId)
						serviceMap[p.ServiceId] = *service
					}
					return serviceMap[p.ServiceId].Name
				},
			},
		}
		dataListTable.AddItems(services)
		common.PrintDataListTable(dataListTable, long)
	},
}

func init() {
	endpointList.Flags().BoolP("long", "l", false, "List additional fields in output")
	endpointList.Flags().StringP("region", "r", "", "Search by region Id")
	endpointList.Flags().StringP("interface", "i", "", "Search by interface")
	endpointList.Flags().StringP("service", "s", "", "Search by service name")

	Endpoint.AddCommand(endpointList)
}