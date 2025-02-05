package keystone

import (
	"github.com/BytemanD/go-console/console"
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
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

		c := common.DefaultClient().KeystoneV3()
		if project == "" {
			users, err := c.User().List(nil)
			utility.LogError(err, "list users failed", true)
			common.PrintUsers(users, long)
		} else {
			users, err := c.ListUsersByProjectId(project)
			if err != nil {
				console.Fatal("get users failed, %s", err)
			}
			common.PrintUsers(users, long)
		}
	},
}

var userShow = &cobra.Command{
	Use:   "show",
	Short: "Show user",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient().KeystoneV3()
		user, err := client.User().Find(args[0])
		utility.LogIfError(err, true, "get user %s failed", args[0])
		common.PrintUser(*user)
	},
}

func init() {
	userList.Flags().BoolP("long", "l", false, "List additional fields in output")
	userList.Flags().String("project", "", "Filter users by project ID")

	User.AddCommand(userList, userShow)
}
