package neutron

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var rule = &cobra.Command{Use: "rule"}

var sgRuleList = &cobra.Command{
	Use:   "list",
	Short: "List security group rules",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient()

		long, _ := cmd.Flags().GetBool("long")
		sgIdOrName, _ := cmd.Flags().GetString("security-group")
		query := url.Values{}
		if sgIdOrName != "" {
			sg, err := c.NeutronV2().SecurityGroup().Find(sgIdOrName)
			utility.LogError(err, "get security group failed", true)
			query.Set("security_group_id", sg.Id)
		}

		sgRules, err := c.NeutronV2().SecurityGroupRule().List(query)
		utility.LogError(err, "list security group failed", true)
		common.PrintSecurityGroupRules(sgRules, long)
	},
}
var sgRuleShow = &cobra.Command{
	Use:   "show <id>",
	Short: "Show security group rule",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient()

		sgRule, err := c.NeutronV2().SecurityGroupRule().Show(args[0])
		utility.LogError(err, "get security group rule failed", true)
		common.PrintSecurityGroupRule(*sgRule)
	},
}

func init() {
	sgRuleList.Flags().BoolP("long", "l", false, "List additional fields in output")
	sgRuleList.Flags().StringP("security-group", "", "", "List according to the project")

	rule.AddCommand(sgRuleList, sgRuleShow)

	group.AddCommand(rule)
	SG.AddCommand(rule)
}
