package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
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
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Enabled", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "DomainId"}, {Name: "Description"},
			},
		}
		pt.AddItems(services)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList)
}
