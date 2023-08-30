package compute

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/BytemanD/stackcrud/common"
	"github.com/BytemanD/stackcrud/openstack/compute"
)

var Aggregate = &cobra.Command{Use: "aggregate"}

var aggList = &cobra.Command{
	Use:   "list",
	Short: "List aggregates",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
		long, _ := cmd.Flags().GetBool("long")
		name, _ := cmd.Flags().GetString("name")

		client := cli.GetClient()
		aggregates, err := client.Compute.AggregateList(nil)
		cli.ExitIfError(err)
		dataTable := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Id"},
				{Name: "Name", Sort: true},
				{Name: "AvailabilityZone"},
			},
			LongColumns: []common.Column{
				{Name: "HostNum", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return len(p.Hosts)
				}},
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return p.MarshalMetadata()
				}},
			},
		}
		filteredAggs := []compute.Aggregate{}
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
		dataTable.AddItems(filteredAggs)
		common.PrintPrettyTable(dataTable, long)
	},
}
var aggShow = &cobra.Command{
	Use:   "show <aggregate id or name>",
	Short: "Show aggregate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agg := args[0]

		client := cli.GetClient()
		aggregate, err := client.Compute.AggregateShow(agg)
		cli.ExitIfError(err)
		dataTable := common.DataTable{
			Item: *aggregate,
			ShortFields: []common.Field{
				{Name: "Id"}, {Name: "Name"}, {Name: "AvailabilityZone"},
				{Name: "Hosts", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return strings.Join(p.Hosts, "\n")
				}},
				{Name: "Metadata", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return p.MarshalMetadata()
				}},
				{Name: "CreatedAt"}, {Name: "UpdatedAt"},
				{Name: "Deleted"}, {Name: "DeletedAt"},
			},
		}
		common.PrintDataTable(dataTable)
	},
}

func init() {
	aggList.Flags().BoolP("long", "l", false, "List additional fields in output")
	aggList.Flags().String("name", "", "List By aggregate name")

	Aggregate.AddCommand(aggList, aggShow)
}
