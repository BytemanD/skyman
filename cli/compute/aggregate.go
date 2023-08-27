package compute

import (
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
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

		dataTable := common.DataListTable{
			ShortHeaders: []string{"Id", "Name", "AvailabilityZone"},
			LongHeaders:  []string{"HostNum", "Metadata"},
			SortBy:       []table.SortBy{{Name: "Name", Mode: table.Asc}},
			Slots: map[string]func(item interface{}) interface{}{
				"HostNum": func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return len(p.Hosts)
				},
				"Metadata": func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return p.MarshalMetadata()
				},
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
		common.PrintDataListTable(dataTable, long)
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
				{Name: "Hosts"}, {Name: "Metadata"},
				{Name: "CreatedAt"}, {Name: "UpdatedAt"},
				{Name: "Deleted"}, {Name: "DeletedAt"},
			},
			Slots: map[string]func(item interface{}) interface{}{
				"Metadata": func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return p.MarshalMetadata()
				},
				"Hosts": func(item interface{}) interface{} {
					p, _ := (item).(compute.Aggregate)
					return strings.Join(p.Hosts, "\n")
				},
			},
		}

		dataTable.Print(false)
	},
}

func init() {
	aggList.Flags().BoolP("long", "l", false, "List additional fields in output")
	aggList.Flags().String("name", "", "List By aggregate name")

	Aggregate.AddCommand(aggList, aggShow)
}
