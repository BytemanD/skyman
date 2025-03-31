package nova

import (
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var (
	groupListFlags flags.GroupListFlags
)

var Group = &cobra.Command{Use: "group"}

var groupList = &cobra.Command{
	Use:   "list",
	Short: "List server groups",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := common.DefaultClient()

		query := url.Values{}

		serverGroups, err := client.NovaV2().ListServerGroup(query)
		utility.LogError(err, "Get server groups failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Name", Sort: true},
				{Name: "Policies", Slot: func(item any) any {
					p, _ := item.(nova.ServerGroup)
					return strings.Join(p.Policies, "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Custom"},
				{Name: "Members", Slot: func(item any) any {
					p, _ := item.(nova.ServerGroup)
					return strings.Join(p.Members, "\n")
				}},
				{Name: "Metadata", Slot: func(item any) any {
					p, _ := item.(nova.ServerGroup)
					return strings.Join(p.GetMetadataList(), "\n")
				}},
				{Name: "ProjectId"},
				{Name: "UserId"},
			},
		}

		pt.AddItems(serverGroups)
		if *groupListFlags.Long {
			pt.StyleSeparateRows = true
		}
		common.PrintPrettyTable(pt, *groupListFlags.Long)

	},
}

func init() {
	groupListFlags = flags.GroupListFlags{
		Long: groupList.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}
	Group.AddCommand(groupList)
	Server.AddCommand(Group)
}
