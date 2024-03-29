package identity

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
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
		c := openstack.DefaultClient().KeystoneV3()
		services, err := c.Services().List(query)
		utility.LogError(err, "list services failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name", Sort: true}, {Name: "Type"},
				{Name: "Enabled", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "Description"},
			},
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	serviceList.Flags().BoolP("long", "l", false, "List additional fields in output")
	serviceList.Flags().StringP("name", "n", "", "Search by service name")
	serviceList.Flags().StringP("type", "t", "", "Search by service type")

	Service.AddCommand(serviceList)
}
