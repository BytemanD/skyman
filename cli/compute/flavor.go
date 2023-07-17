package compute

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/stackcrud/cli"
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

		flavors, err := client.Compute.FlavorListDetail(nil)
		if err != nil {
			logging.Fatal("%s", err)
		}
		if long {
			for i, flavor := range flavors {
				extraSpecs, err := client.Compute.FlavorExtraSpecsList(flavor.Id)
				if err != nil {
					logging.Fatal("get flavor extra specs failed %s", err)
				}
				flavors[i].ExtraSpecs = extraSpecs
			}
		}
		flavors.PrintTable(long)

	},
}

func init() {
	// flavor list flags
	flavorList.Flags().Bool("public", false, "List public flavors")
	flavorList.Flags().BoolP("long", "l", false, "List additional fields in output")
	// Server create flags

	Flavor.AddCommand(flavorList)
}
