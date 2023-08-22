package identity

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/cli"
)

var Project = &cobra.Command{Use: "project"}

var projectList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")

		client := cli.GetClient()
		services, err := client.Identity.ProjectList(nil)
		if err != nil {
			logging.Fatal("get users failed, %s", err)
		}
		dataTable := cli.DataListTable{
			ShortHeaders: []string{"Id", "Name"},
			LongHeaders:  []string{"DomainId", "Description", "Enabled"},
			SortBy:       []table.SortBy{{Name: "Region"}, {Name: "Service Name"}},
		}
		dataTable.AddItems(services)
		dataTable.Print(long)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList)
}
