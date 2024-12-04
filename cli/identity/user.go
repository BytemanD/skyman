package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var User = &cobra.Command{Use: "user"}

var userList = &cobra.Command{
	Use:   "list",
	Short: "List users",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		project, _ := cmd.Flags().GetString("project")

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

		c := openstack.DefaultClient().KeystoneV3()
		if project == "" {
			users, err := c.Users().List(nil)
			utility.LogError(err, "list users failed", true)
			pt.AddItems(users)
		} else {
			users, err := c.Users().ListByProjectId(project)
			if err != nil {
				logging.Fatal("get users failed, %s", err)
			}
			pt.AddItems(users)
		}
		common.PrintPrettyTable(pt, long)
	},
}

var userShow = &cobra.Command{
	Use:   "show",
	Short: "Show user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		client := openstack.DefaultClient().KeystoneV3()
		user, err := client.Users().User().Found(args[0])
		utility.LogIfError(err, true, "get user %s failed", args[0])
		pt := common.PrettyItemTable{
			Item: *user,
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
	userList.Flags().BoolP("long", "l", false, "List additional fields in output")
	userList.Flags().String("project", "", "Filter users by project ID")

	User.AddCommand(userList, userShow)
}
