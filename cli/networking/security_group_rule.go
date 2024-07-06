package networking

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var rule = &cobra.Command{Use: "rule"}

var sgRuleList = &cobra.Command{
	Use:   "list",
	Short: "List security groups",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient()

		long, _ := cmd.Flags().GetBool("long")
		sgIdOrName, _ := cmd.Flags().GetString("security-group")
		query := url.Values{}
		if sgIdOrName != "" {
			sg, err := c.NeutronV2().SecurityGroups().Found(sgIdOrName)
			utility.LogError(err, "get security group failed", true)
			query.Set("security_group_id", sg.Id)
		}

		sgRules, err := c.NeutronV2().SecurityGroupRules().List(query)
		utility.LogError(err, "list security group failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Protocol"},
				{Name: "RemoteIpPrefix"},
				{Name: "PortRange", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.SecurityGroupRule)
					return p.PortRange()
				}},
				{Name: "RemoteGroupId"},
				{Name: "SecurityGroupId"},
			},
			LongColumns: []common.Column{
				{Name: "Direction"},
				{Name: "Ethertype"},
			},
		}
		pt.AddItems(sgRules)
		common.PrintPrettyTable(pt, long)
	},
}
var sgRuleShow = &cobra.Command{
	Use:   "show <id>",
	Short: "Show security group rule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := openstack.DefaultClient()

		sg, err := c.NeutronV2().SecurityGroupRules().Show(args[0])
		utility.LogError(err, "get security group rule failed", true)
		pt := common.PrettyItemTable{
			Item: *sg,
			ShortFields: []common.Column{
				{Name: "Id"},
				{Name: "Protocol"},
				{Name: "Direction"},
				{Name: "Ethertype"},
				{Name: "RemoteIpPrefix"},
				{Name: "PortRange", Slot: func(item interface{}) interface{} {
					p, _ := item.(neutron.SecurityGroupRule)
					return p.PortRange()
				}},
				{Name: "RemoteGroupId"},
				{Name: "SecurityGroupId"},
				{Name: "RevisionNumber"},
				{Name: "CreatedAt"},
				{Name: "UpdatedAt"},
				{Name: "ProjectId"},
			},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	sgRuleList.Flags().BoolP("long", "l", false, "List additional fields in output")
	sgRuleList.Flags().StringP("security-group", "", "", "List according to the project")

	rule.AddCommand(sgRuleList, sgRuleShow)

	group.AddCommand(rule)
}
