package compute

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var Aggregate = &cobra.Command{Use: "aggregate"}

var aggList = &cobra.Command{
	Use:   "list",
	Short: "List aggregates",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")

		client := openstack.DefaultClient()
		aggregates, err := client.NovaV2().Aggregate().List(nil)
		utility.LogError(err, "list aggregates failed", true)
		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Name", Sort: true},
				{Name: "AvailabilityZone"},
				{Name: "HostNum", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Aggregate)
					return len(p.Hosts)
				}},
			},
			LongColumns: []common.Column{
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Aggregate)
					return p.MarshalMetadata()
				}},
			},
		}
		filteredAggs := []nova.Aggregate{}
		if name != "" {
			for _, agg := range aggregates {
				if !strings.Contains(agg.Name, name) {
					continue
				}
				filteredAggs = append(filteredAggs, agg)
			}
		} else {
			filteredAggs = aggregates
		}
		pt.AddItems(filteredAggs)
		common.PrintPrettyTable(pt, long)
	},
}
var aggShow = &cobra.Command{
	Use:   "show <aggregate id or name>",
	Short: "Show aggregate",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		agg := args[0]

		client := openstack.DefaultClient()
		aggregate, err := client.NovaV2().Aggregate().Show(agg)
		utility.LogError(err, "show aggregate failed", true)
		pt := common.PrettyItemTable{
			Item: *aggregate,
			ShortFields: []common.Column{
				{Name: "Id"}, {Name: "Name"}, {Name: "AvailabilityZone"},
				{Name: "Hosts", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Aggregate)
					return strings.Join(p.Hosts, "\n")
				}},
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.Aggregate)
					return p.MarshalMetadata()
				}},
				{Name: "CreatedAt"}, {Name: "UpdatedAt"},
				{Name: "Deleted"}, {Name: "DeletedAt"},
			},
		}
		common.PrintPrettyItemTable(pt)
	},
}

func init() {
	aggList.Flags().BoolP("long", "l", false, "List additional fields in output")
	aggList.Flags().String("name", "", "List By aggregate name")

	Aggregate.AddCommand(aggList, aggShow)
}
