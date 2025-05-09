package nova

import (
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
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
		client := common.DefaultClient()
		aggregates, err := client.NovaV2().ListAgg(nil)
		utility.LogError(err, "list aggregates failed", true)
		if *aggListFlags.Name != "" {
			aggregates = lo.Filter(aggregates, func(item nova.Aggregate, _ int) bool {
				return strings.Contains(item.Name, *aggListFlags.Name)
			})
		}

		common.PrintAggregates(aggregates, *aggListFlags.Long)
	},
}
var aggShow = &cobra.Command{
	Use:   "show <aggregate id or name>",
	Short: "Show aggregate",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		aggregate, err := client.NovaV2().FindAgg(args[0])
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
		client := common.DefaultClient()
		aggregate, err := client.NovaV2().CreateAgg(agg)
		utility.LogIfError(err, true, "create aggregate %s failed", name)
		common.PrintAggregate(*aggregate)
	},
}
var aggDelete = &cobra.Command{
	Use:   "delete <aggregate> [<aggregate> ...]",
	Short: "delete aggregate(s)",
	Args:  cobra.MinimumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		client := common.DefaultClient()
		for _, agg := range args {
			aggregate, err := client.NovaV2().FindAgg(agg)
			utility.LogIfError(err, true, "get aggregate %s failed", agg)
			err = client.NovaV2().DeleteAgg(aggregate.Id)
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
		client := common.DefaultClient()
		aggregate, err := client.NovaV2().FindAgg(idOrName)
		utility.LogIfError(err, true, "get aggregate %s failed", idOrName)
		added := 0
		for _, host := range hosts {
			agg, err := client.NovaV2().AggAddHost(aggregate.Id, host)
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
		client := common.DefaultClient()
		aggregate, err := client.NovaV2().FindAgg(idOrName)
		utility.LogIfError(err, true, "get aggregate %s failed", idOrName)
		for _, host := range hosts {
			console.Debug("remove host %s from aggregate %s", host, idOrName)
			agg, err := client.NovaV2().AggRemoveHost(aggregate.Id, host)
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
