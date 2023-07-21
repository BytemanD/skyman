package identity

import (
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

var Token = &cobra.Command{Use: "token"}

var tokenIssue = &cobra.Command{
	Use:   "issue <server>",
	Short: "Issue new token",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := cli.GetClient()
		if err != nil {
			logging.Fatal("get openstack client failed %s", err)
		}
		token := client.Identity.GetToken()

		dataTable := cli.DataTable{
			Item: token,
			ShortFields: []cli.Field{
				{Name: "ExpiresAt", Text: "Expires At"},
				{Name: "TokenId", Text: "Id"},
				{Name: "ProjectId", Text: "Project Id"},
				{Name: "UserId", Text: "User Id"},
			},
			Slots: map[string]func(item interface{}) interface{}{
				"ProjectId": func(item interface{}) interface{} {
					p, _ := (item).(identity.Token)
					return p.Project.Id
				},
				"UserId": func(item interface{}) interface{} {
					p, _ := (item).(identity.Token)
					return p.User.Id
				},
			},
		}
		dataTable.Print(false)
	},
}

func init() {
	Token.AddCommand(tokenIssue)
}
