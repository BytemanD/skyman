package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
)

var Project = &cobra.Command{Use: "project"}

var projectList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")

		client := cli.GetClient()
		projects, err := client.Identity.ProjectList(nil)
		if err != nil {
			logging.Fatal("get projects failed, %s", err)
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
		pt.AddItems(projects)
		common.PrintPrettyTable(pt, long)
	},
}
var projectDelete = &cobra.Command{
	Use:   "delete <project id>",
	Short: "Delete endpoints",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		err := client.Identity.ProjectDelete(args[0])
		common.LogError(err, "delete project failed", true)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList, projectList)
}
