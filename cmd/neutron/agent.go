package neutron

import (
	"net/url"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/utility"
)

var agentCmd = &cobra.Command{Use: "agent", Short: "Network Agent comamnds"}

var agentList = &cobra.Command{
	Use:   "list",
	Short: "List agent",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		c := openstack.DefaultClient().NeutronV2()

		long, _ := cmd.Flags().GetBool("long")
		binary, _ := cmd.Flags().GetString("binary")
		host, _ := cmd.Flags().GetString("host")
		query := url.Values{}
		if binary != "" {
			query.Set("binary", binary)
		}
		if host != "" {
			query.Set("host", host)
		}
		agents, err := c.Agent().List(query)
		utility.LogError(err, "list ports failed", true)
		common.PrintAgents(agents, long)
	},
}

func init() {
	agentList.Flags().String("host", "", "filter by host")
	agentList.Flags().String("binary", "", "filter by binary")
	agentCmd.AddCommand(agentList)
	Network.AddCommand(agentCmd)
}
