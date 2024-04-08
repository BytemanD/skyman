package networking

import (
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/neutron"
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
		agents, err := c.Agents().List(query)
		utility.LogError(err, "list ports failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"}, {Name: "AgentType"},
				{Name: "Host"},
				{Name: "AvailabilityZone"},
				{Name: "Alive", AutoColor: true, Slot: func(item interface{}) interface{} {
					p := item.(neutron.Agent)
					if p.Alive {
						return ":-)"
					}
					return "XXX"
				}},
				{Name: "AdminStateUp", Text: "State"},
				{Name: "Binary"},
			},
			ColumnConfigs: []table.ColumnConfig{{Number: 4, Align: text.AlignRight}},
		}
		pt.AddItems(agents)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	agentCmd.AddCommand(agentList)
	Network.AddCommand(agentCmd)
}
