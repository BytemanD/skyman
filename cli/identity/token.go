package identity

import (
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/utility"
)

var Token = &cobra.Command{Use: "token"}

var tokenIssue = &cobra.Command{
	Use:   "issue <server>",
	Short: "Issue new token",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := openstack.DefaultClient()

		tokenId, err := client.AuthPlugin.GetTokenId()
		utility.LogError(err, "get token failed", true)
		token, _ := client.AuthPlugin.GetToken()
		pt := common.PrettyItemTable{
			Item:            *token,
			Number2WidthMax: 184,
			ShortFields: []common.Column{
				{Name: "ExpiresAt", Text: "Expires At"},
				{Name: "TokenId", Text: "Id", Slot: func(item interface{}) interface{} {
					return tokenId
				}},
				{Name: "ProjectId", Text: "Project Id", Slot: func(item interface{}) interface{} {
					p, _ := (item).(auth.Token)
					return p.Project.Id
				}},
				{Name: "UserId", Text: "User Id", Slot: func(item interface{}) interface{} {
					p, _ := (item).(auth.Token)
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
