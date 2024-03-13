package compute

import (
	"net/url"
	"strings"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Group = &cobra.Command{Use: "group"}

var groupList = &cobra.Command{
	Use:   "list",
	Short: "List server groups",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		client := cli.GetClient()

		query := url.Values{}

		long, _ := cmd.Flags().GetBool("long")
		serverGroups, err := client.ComputeClient().ServerGroupList(query)
		utility.LogError(err, "Get server groups failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Name", Sort: true},
				{Name: "Policies", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.ServerGroup)
					return strings.Join(p.Policies, "\n")
				}},
			},
			LongColumns: []common.Column{
				{Name: "Custom"},
				{Name: "Members", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.ServerGroup)
					return strings.Join(p.Members, "\n")
				}},
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					p, _ := item.(compute.ServerGroup)
					return strings.Join(p.GetMetadataList(), "\n")
				}},
				{Name: "ProjectId"},
				{Name: "UserId"},
			},
		}

		pt.AddItems(serverGroups)
		if long {
			pt.StyleSeparateRows = true
		}
		common.PrintPrettyTable(pt, long)

	},
}

func init() {
	common.RegistryLongFlag(groupList)

	Group.AddCommand(groupList)

	Server.AddCommand(Group)
}
