package compute

import (
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Flavor = &cobra.Command{Use: "flavor"}

var flavorList = &cobra.Command{
	Use:   "list",
	Short: "List flavors",
	Run: func(cmd *cobra.Command, _ []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}

		query := url.Values{}
		public, _ := cmd.Flags().GetBool("public")
		if public {
			query.Set("public", "true")
		}
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")

		flavors, err := client.Compute.FlavorListDetail(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
		filteredFlavors := compute.Flavors{}
		if name != "" {
			for _, flavor := range flavors {
				if strings.Contains(flavor.Name, name) {
					filteredFlavors = append(filteredFlavors, flavor)
				}
			}
		} else {
			filteredFlavors = flavors
		}
		if long {
			for i, flavor := range filteredFlavors {
				extraSpecs, err := client.Compute.FlavorExtraSpecsList(flavor.Id)
				if err != nil {
					logging.Fatal("get flavor extra specs failed %s", err)
				}
				flavors[i].ExtraSpecs = extraSpecs
			}
		}
		filteredFlavors.PrintTable(long)
	},
}

func init() {
	// flavor list flags
	flavorList.Flags().Bool("public", false, "List public flavors")
	flavorList.Flags().StringP("name", "n", "", "Show flavors matched by name")
	flavorList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Flavor.AddCommand(flavorList)
}
