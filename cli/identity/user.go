package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
)

var User = &cobra.Command{Use: "user"}

var userList = &cobra.Command{
	Use:   "list",
	Short: "List endpoints",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		project, _ := cmd.Flags().GetString("project")

		client := cli.GetClient()
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "Name"},
				{Name: "Enabled", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "Project"}, {Name: "DomainId"},
				{Name: "Description"}, {Name: "Email"},
			},
		}

		if project == "" {
			users, err := client.Identity.UserList(nil)
			if err != nil {
				logging.Fatal("get users failed, %s", err)
			}
			pt.AddItems(users)
		} else {
			users, err := client.Identity.UserListByProjectId(project)
			if err != nil {
				logging.Fatal("get users failed, %s", err)
			}
			pt.AddItems(users)
		}
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	userList.Flags().BoolP("long", "l", false, "List additional fields in output")
	userList.Flags().String("project", "", "Filter users by project ID")

	User.AddCommand(userList)
}
