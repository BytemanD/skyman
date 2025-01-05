package keystone

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/keystone"
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
		services, err := c.Service().List(query)
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
var serviceCreate = &cobra.Command{
	Use:   "create <type>",
	Short: "Create service",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serviceType := args[0]

		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		disable, _ := cmd.Flags().GetBool("disable")

		c := openstack.DefaultClient().KeystoneV3()

		service, err := c.Service().Create(
			keystone.Service{
				Type:     serviceType,
				Enabled:  !disable,
				Resource: model.Resource{Name: name, Description: description},
			})
		utility.LogError(err, "create service failed", true)

		pt := common.PrettyItemTable{
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "Type"},
				{Name: "Enabled"},
				{Name: "Description"},
			},
			LongFields: []common.Column{},
			Item:       *service,
		}
		common.PrintPrettyItemTable(pt)
	},
}
var serviceDelete = &cobra.Command{
	Use:   "delete <service> [service ...]",
	Short: "Delete service",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().KeystoneV3()

		for _, id := range args {
			console.Info("request to delete service %s", id)
			err := c.Service().Delete(id)
			utility.LogError(err, fmt.Sprintf("delete service %s failed", id), false)
		}
	},
}

func init() {
	serviceList.Flags().BoolP("long", "l", false, "List additional fields in output")
	serviceList.Flags().StringP("name", "n", "", "Search by service name")
	serviceList.Flags().StringP("type", "t", "", "Search by service type")

	serviceCreate.Flags().Bool("disable", false, "Disable service")
	serviceCreate.Flags().StringP("name", "n", "", "New service name")
	serviceCreate.Flags().StringP("description", "t", "", "New service description")

	Service.AddCommand(serviceList, serviceCreate, serviceDelete)
}
