package identity

import (
	"github.com/jedib0t/go-pretty/v6/table"
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

		dataListTable := common.DataListTable{
			ShortHeaders: []string{"Id", "Name"},
			LongHeaders:  []string{"Project", "DomainId", "Description", "Email", "Enabled"},
			SortBy:       []table.SortBy{{Name: "Region"}, {Name: "Service Name"}},
		}
		if project == "" {
			users, err := client.Identity.UserList(nil)
			if err != nil {
				logging.Fatal("get users failed, %s", err)
			}
			dataListTable.AddItems(users)
		} else {
			users, err := client.Identity.UserListByProjectId(project)
			if err != nil {
				logging.Fatal("get users failed, %s", err)
			}
			dataListTable.AddItems(users)
		}
		common.PrintDataListTable(dataListTable, long)
	},
}

func init() {
	userList.Flags().BoolP("long", "l", false, "List additional fields in output")
	userList.Flags().String("project", "", "Filter users by project ID")

	User.AddCommand(userList)
}
