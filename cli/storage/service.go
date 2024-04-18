package storage

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/spf13/cobra"
)

var service = &cobra.Command{Use: "service"}
var list = &cobra.Command{
	Use:   "list",
	Short: "List volume services",
	Run: func(cmd *cobra.Command, _ []string) {
		client := openstack.DefaultClient()

		query := url.Values{}
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")
		zone, _ := cmd.Flags().GetString("zone")
		long, _ := cmd.Flags().GetBool("long")

		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}

		services, _ := client.CinderV2().Services().List(query)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Binary"},
				{Name: "Host"}, {Name: "Zone"},
				{Name: "Status", AutoColor: true},
				{Name: "State", AutoColor: true},
				{Name: "UpdatedAt"},
			},
			LongColumns: []common.Column{
				{Name: "DisabledReason"},
			},
			Filters: map[string]string{},
		}
		if zone != "" {
			pt.ShortColumns[3].Filters = []string{zone}
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	list.Flags().BoolP("long", "l", false, "List additional fields in output")

	service.AddCommand(list)

	Volume.AddCommand(service)
}
