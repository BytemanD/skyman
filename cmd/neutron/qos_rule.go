package neutron

import (
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var qosRule = &cobra.Command{Use: "rule"}

var qosRuleList = &cobra.Command{
	Use:   "list <qos-policy>",
	Short: "list qos rules",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient()

		policy, err := c.NeutronV2().FindQosPolicy(args[0])
		utility.LogIfError(err, true, "get qos policy %s failed", args[0])

		common.PrintQosPolicyRules(policy.Rules, false)
	},
}

func init() {
	qosRule.AddCommand(qosRuleList)
	Qos.AddCommand(qosRule)
}
