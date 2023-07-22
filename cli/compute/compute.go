package compute

import (
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
)

var Compute = &cobra.Command{Use: "compute"}
var computeService = &cobra.Command{Use: "service"}

var csList = &cobra.Command{
	Use:   "list",
	Short: "List compute services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		client.Compute.UpdateVersion()
		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")

		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}
		services := client.Compute.ServiceList(query)
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Binary", "Host", "Zone", "Status", "State", "UpdatedAt"},
			LongHeaders: []string{
				"DisabledReason", "ForcedDown"},
			HeaderLabel: map[string]string{
				"UpdatedAt":      "Update At",
				"DisabledReason": "Disabled Reason",
				"ForcedDown":     "Forced Down",
			},
			SortBy: []table.SortBy{
				{Name: "Name", Mode: table.Asc},
			},
		}
		for _, service := range services {
			dataTable.Items = append(dataTable.Items, service)
		}
		dataTable.Print(long)
	},
}

func init() {
	// compute service
	csList.Flags().String("binary", "", "Search by binary")
	csList.Flags().String("host", "", "Search by hostname")
	csList.Flags().StringArrayP("state", "s", nil, "Search by server status")
	csList.Flags().BoolP("long", "l", false, "List additional fields in output")
	computeService.AddCommand(csList)

	Compute.AddCommand(computeService)
}
