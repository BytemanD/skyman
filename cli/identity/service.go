package identity

import (
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
)

var Service = &cobra.Command{Use: "service"}

var serviceList = &cobra.Command{
	Use:   "list",
	Short: "List services",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		serviceName, _ := cmd.Flags().GetString("name")
		serviceType, _ := cmd.Flags().GetString("type")
		query := url.Values{}
		if serviceName != "" {
			query.Add("name", serviceName)
		}
		if serviceType != "" {
			query.Add("type", serviceType)
		}
		client := cli.GetClient()
		services, err := client.Identity.ServiceList(query)
		if err != nil {
			logging.Fatal("get services failed, %s", err)
		}
		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Id", "Name", "Type", "Enabled"},
			LongHeaders:  []string{"Description"},
			SortBy:       []table.SortBy{{Name: "Name", Mode: table.Asc}},
		}
		dataListTable.AddItems(services)
		common.PrintDataListTable(dataListTable, long)
	},
}

func init() {
	serviceList.Flags().BoolP("long", "l", false, "List additional fields in output")
	serviceList.Flags().StringP("name", "n", "", "Search by service name")
	serviceList.Flags().StringP("type", "t", "", "Search by service type")

	Service.AddCommand(serviceList)
}
