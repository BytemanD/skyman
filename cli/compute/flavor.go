package compute

import (
	"net/url"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
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
		dataTable := cli.DataListTable{
			ShortHeaders: []string{
				"Id", "Name", "Vcpus", "Ram", "Disk", "Ephemeral", "IsPublic"},
			LongHeaders: []string{
				"Swap", "RXTXFactor", "ExtraSpecs"},
			HeaderLabel: map[string]string{
				"IsPublic":   "Is Public",
				"RXTXFactor": "RXTX Factor",
				"ExtraSpecs": "Extra Specs",
			},
			SortBy: []table.SortBy{{Name: "Name", Mode: table.Asc}},
			Slots: map[string]func(item interface{}) interface{}{
				"ExtraSpecs": func(item interface{}) interface{} {
					p, _ := (item).(compute.Flavor)
					return strings.Join(p.ExtraSpecs.GetList(), "\n")
				},
			},
		}
		if long {
			for i, flavor := range filteredFlavors {
				extraSpecs, err := client.Compute.FlavorExtraSpecsList(flavor.Id)
				if err != nil {
					logging.Fatal("get flavor extra specs failed %s", err)
				}
				filteredFlavors[i].ExtraSpecs = extraSpecs
			}
		}
		for _, flavor := range filteredFlavors {
			dataTable.Items = append(dataTable.Items, flavor)
		}
		dataTable.Print(long)
	},
}

func init() {
	// flavor list flags
	flavorList.Flags().Bool("public", false, "List public flavors")
	flavorList.Flags().StringP("name", "n", "", "Show flavors matched by name")
	flavorList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Flavor.AddCommand(flavorList)
}
