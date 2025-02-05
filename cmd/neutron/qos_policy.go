package neutron

import (
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/utility"
	"github.com/spf13/cobra"
)

var Qos = &cobra.Command{Use: "qos"}

var policy = &cobra.Command{Use: "policy"}

var qosPolicyList = &cobra.Command{
	Use:   "list",
	Short: "List qos policies",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := common.DefaultClient()

		// long, _ := cmd.Flags().GetBool("long")
		projectIdOrName, _ := cmd.Flags().GetString("project")
		query := url.Values{}

		if projectIdOrName != "" {
			project, err := c.KeystoneV3().Project().Find(projectIdOrName)
			utility.LogError(err, "get project failed", true)
			query.Set("project_id", project.Id)
		}

		policies, err := c.NeutronV2().QosPolicy().List(query)
		utility.LogError(err, "list qos policy failed", true)
		common.PrintQosPolicys(policies, false)
	},
}
var qosPolicyShow = &cobra.Command{
	Use:   "show <qos-policy>",
	Short: "Show qos policy",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		c := common.DefaultClient()

		policy, err := c.NeutronV2().QosPolicy().Find(args[0])
		utility.LogIfError(err, true, "get qos policy %s failed", args[0])
		common.PrintQosPolicy(*policy)
	},
}

func init() {
	// qosPolicyList.Flags().BoolP("long", "l", false, "List additional fields in output")
	qosPolicyList.Flags().StringP("project", "", "", "List according to the project")

	policy.AddCommand(qosPolicyList, qosPolicyShow)
	Qos.AddCommand(policy)

	Network.AddCommand(Qos)
}
