package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var Project = &cobra.Command{Use: "project"}

var projectList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")

		c := openstack.DefaultClient().KeystoneV3()
		projects, err := c.Projects().List(nil)
		utility.LogError(err, "list projects failed", true)

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
	Short: "Delete project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().KeystoneV3()
		err := c.Projects().Delete(args[0])
		utility.LogError(err, "delete project failed", true)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList, projectDelete)
}
