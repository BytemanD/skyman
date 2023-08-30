package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/identity"
)

var Token = &cobra.Command{Use: "token"}

var tokenIssue = &cobra.Command{
	Use:   "issue <server>",
	Short: "Issue new token",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()

		token := client.AuthClient.GetToken()
		pt := common.PrettyItemTable{
			Item: token,
			ShortFields: []common.Column{
				{Name: "ExpiresAt", Text: "Expires At"},
				{Name: "TokenId", Text: "Id"},
				{Name: "ProjectId", Text: "Project Id", Slot: func(item interface{}) interface{} {
					p, _ := (item).(identity.Token)
					return p.Project.Id
				}},
				{Name: "UserId", Text: "User Id", Slot: func(item interface{}) interface{} {
					p, _ := (item).(identity.Token)
					return p.User.Id
				}},
			},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	Token.AddCommand(tokenIssue)
}
