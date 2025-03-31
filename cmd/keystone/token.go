package keystone

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/datatable"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/utility"
)

var Token = &cobra.Command{Use: "token"}

var tokenIssue = &cobra.Command{
	Use:   "issue <server>",
	Short: "Issue new token",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		client := common.DefaultClient()

		token, err := client.AuthPlugin.GetToken()
		if err != nil {
			utility.LogError(err, "get token issue failed", true)
		}
		common.PrintItem(
			[]datatable.Field[model.Token]{
				{Name: "TokenId", Text: "Id", RenderFunc: func(item model.Token) any {
					return token.TokenId
				}},
				{Name: "ExpiresAt"},
				{Name: "ProjectId", Text: "Project", RenderFunc: func(item model.Token) any {
					return fmt.Sprintf("%s (%s)", item.Project.Id, item.Project.Name)
				}},
				{Name: "UserId", Text: "User", RenderFunc: func(item model.Token) any {
					return fmt.Sprintf("%s (%s)", item.User.Id, item.User.Name)
				}},
				{Name: "Roles", RenderFunc: func(item model.Token) any {
					return strings.Join(client.AuthPlugin.Roles(), ", ")
				}},
			},
			[]datatable.Field[model.Token]{},
			*token, common.TableOptions{ValueColumnMaxWidth: 184},
		)
	},
}

func init() {
	Token.AddCommand(tokenIssue)
}
