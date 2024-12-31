package nova

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/cli/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	aggListFlags   flags.AggregateListFlags
	aggCreateFlags flags.AggregateCreateFlags
)

var Aggregate = &cobra.Command{Use: "aggregate"}

var aggList = &cobra.Command{
	Use:   "list",
	Short: "List aggregates",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, _ []string) {
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
		if *aggListFlags.Name != "" {
			for _, agg := range aggregates {
				if !strings.Contains(agg.Name, *aggListFlags.Name) {
					continue
				}
				filteredAggs = append(filteredAggs, agg)
			}
		} else {
			filteredAggs = aggregates
		}
		pt.AddItems(filteredAggs)
		common.PrintPrettyTable(pt, *aggListFlags.Long)
	},
}
var aggShow = &cobra.Command{
	Use:   "show <aggregate id or name>",
	Short: "Show aggregate",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		aggregate, err := client.NovaV2().Aggregate().Find(args[0])
		utility.LogIfError(err, true, "get aggregate %s failed", args[0])
		common.PrintAggregate(*aggregate)
	},
}
var aggCreate = &cobra.Command{
	Use:   "create <name>",
	Short: "create aggregate",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		agg := nova.Aggregate{Name: name}
		if *aggCreateFlags.AZ != "" {
			agg.AvailabilityZone = *aggCreateFlags.AZ
		}
		client := openstack.DefaultClient()
		aggregate, err := client.NovaV2().Aggregate().Create(agg)
		utility.LogIfError(err, true, "create aggregate %s failed", name)
		common.PrintAggregate(*aggregate)
	},
}
var aggDelete = &cobra.Command{
	Use:   "delete <aggregate> [<aggregate> ...]",
	Short: "delete aggregate(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := openstack.DefaultClient()
		for _, agg := range args {
			aggregate, err := client.NovaV2().Aggregate().Find(agg)
			utility.LogIfError(err, true, "get aggregate %s failed", agg)
			err = client.NovaV2().Aggregate().Delete(aggregate.Id)
			utility.LogIfError(err, true, "delete aggregate %s failed", agg)
		}
	},
}
var aggAdd = &cobra.Command{Use: "add"}
var aggRemove = &cobra.Command{Use: "remove"}
var addHost = &cobra.Command{
	Use:   "host <aggregate> <host1> [<host2>...]",
	Short: "Add hosts to aggregate",
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		idOrName, hosts := args[0], args[1:]
		client := openstack.DefaultClient()
		aggregate, err := client.NovaV2().Aggregate().Find(idOrName)
		utility.LogIfError(err, true, "get aggregate %s failed", idOrName)
		added := 0
		for _, host := range hosts {
			agg, err := client.NovaV2().Aggregate().AddHost(aggregate.Id, host)
			utility.LogIfError(err, false, "add %s to aggregate %s failed", host, idOrName)
			if err == nil {
				aggregate = agg
			} else {
				added += 1
			}
		}
		if added != 0 {
			common.PrintAggregate(*aggregate)
		}
	},
}
var removeHost = &cobra.Command{
	Use:   "host <aggregate> <host1> [<host2>...]",
	Short: "Add hosts to aggregate",
	Args:  cobra.MinimumNArgs(2),
	Run: func(_ *cobra.Command, args []string) {
		idOrName, hosts := args[0], args[1:]
		client := openstack.DefaultClient()
		aggregate, err := client.NovaV2().Aggregate().Find(idOrName)
		utility.LogIfError(err, true, "get aggregate %s failed", idOrName)
		for _, host := range hosts {
			logging.Debug("remove host %s from aggregate %s", host, idOrName)
			agg, err := client.NovaV2().Aggregate().RemoveHost(aggregate.Id, host)
			utility.LogIfError(err, false, "remove %s to aggregate %s failed", host, idOrName)
			if err == nil {
				aggregate = agg
			}
		}
		common.PrintAggregate(*aggregate)
	},
}

func init() {
	aggListFlags = flags.AggregateListFlags{
		Long: aggList.Flags().BoolP("long", "l", false, "List additional fields in output"),
		Name: aggList.Flags().String("name", "", "List By aggregate name"),
	}
	aggCreateFlags = flags.AggregateCreateFlags{
		AZ: aggCreate.Flags().String("az", "", "The availability zone of the aggregate"),
	}

	aggAdd.AddCommand(addHost)
	aggRemove.AddCommand(removeHost)
	Aggregate.AddCommand(aggList, aggShow, aggCreate, aggDelete, aggAdd, aggRemove)
}
