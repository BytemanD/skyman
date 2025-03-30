package keystone

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
)

var Project = &cobra.Command{Use: "project"}

var projectList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")

		c := common.DefaultClient().KeystoneV3()
		projects, err := c.ListProject(nil)
		utility.LogError(err, "list projects failed", true)
		common.PrintProjects(projects, long)
	},
}
var projectDelete = &cobra.Command{
	Use:   "delete <project id>",
	Short: "Delete project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().KeystoneV3()
		err := c.DeleteProject(args[0])
		utility.LogError(err, "delete project failed", true)
	},
}
var projectShow = &cobra.Command{
	Use:   "show <id or name>",
	Short: "Show project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient().KeystoneV3()
		project, err := c.FindProject(args[0])
		utility.LogError(err, "show project failed", true)
		common.PrintProject(*project)
	},
}

func init() {
	projectList.Flags().BoolP("long", "l", false, "List additional fields in output")

	Project.AddCommand(projectList, projectShow, projectDelete)
}
