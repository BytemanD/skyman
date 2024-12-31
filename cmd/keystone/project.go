package keystone

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
		projects, err := c.Project().List(nil)
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
		err := c.Project().Delete(args[0])
		utility.LogError(err, "delete project failed", true)
	},
}
var projectShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient().KeystoneV3()
		project, err := c.Project().Find(args[0])
		utility.LogError(err, "show project failed", true)
		pt := common.PrettyItemTable{
			Item: *project,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Description"},
				{Name: "DomainId"},
				{Name: "Enabled", AutoColor: true},
				{Name: "IsDomain"},
				{Name: "ParentId"},
				{Name: "Tags"},
			},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList, projectShow, projectDelete)
}
