package nova

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/cmd/flags"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

var (
	serverActionFlags flags.ServerActionFlags
)

func listServerActions(serverId string, actionName string, last int, long bool) {
	client := common.DefaultClient()

	actions, err := client.NovaV2().Server().ListActions(serverId)
	utility.LogError(err, "list actions failed", true)
	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "Action"}, {Name: "RequestId"}, {Name: "StartTime", Sort: true},
			{Name: "Message"},
		},
		LongColumns: []common.Column{
			{Name: "ProjectId"}, {Name: "UserId"},
		},
	}
	if actionName != "" {
		actions = lo.Filter(actions, func(item nova.InstanceAction, _ int) bool {
			return item.Action == actionName
		})
	}
	if last > 0 {
		actions = common.LastN(actions, last)
	}
	pt.Items = append(pt.Items, actions)
	common.PrintPrettyTable(pt, long)
}
func listServerActionsWithSpend(serverId string, actionName string, requestId string, last int, long bool) {
	client := common.DefaultClient()
	actionsWithEvents, err := client.NovaV2().Server().ListActionsWithEvents(
		serverId, actionName, requestId, last)
	utility.LogError(err, "get server actions and events failed", true)

	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "Name", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.InstanceAction)
				return p.Action
			}},
			{Name: "RequestId", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.InstanceAction)
				return p.RequestId
			}},
			{Name: "StartTime", Slot: func(item interface{}) interface{} {
				p, _ := item.(nova.InstanceAction)
				return p.StartTime
			}},
			{Name: "SpendTime", Text: "Spend time (seconds)",
				Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.InstanceAction)
					spendTime, err := p.GetSpendTime()
					if err != nil {
						return err
					} else {
						return fmt.Sprintf("%.3f", spendTime)
					}
				}},
		},
	}
	for _, actionWithEvents := range actionsWithEvents {
		for _, event := range actionWithEvents.Events {
			columnIndex := pt.GetLongColumnIndex(event.Event)
			if columnIndex >= 0 {
				continue
			}
			pt.LongColumns = append(pt.LongColumns,
				common.Column{
					Name: event.Event,
					SlotColumn: func(item interface{}, column common.Column) interface{} {
						p, _ := (item).(nova.InstanceAction)
						for _, e := range p.Events {
							if e.Event == column.Name {
								spendTime, err := e.GetSpendTime()
								if err != nil {
									return err
								} else {
									return fmt.Sprintf("%.3f", spendTime)
								}
							}
						}
						return "-"
					},
				},
			)
		}
	}
	pt.AddItems(actionsWithEvents)
	common.PrintPrettyTable(pt, long)
}
func showAction(serverId string, requestId string, long bool) {
	client := common.DefaultClient()

	action, err := client.NovaV2().Server().ShowAction(serverId, requestId)
	utility.LogError(err, "get server action failed", true)
	pt := common.PrettyTable{
		Title: fmt.Sprintf("Action: %s", action.Action),
		ShortColumns: []common.Column{
			{Name: "Event"}, {Name: "Host"},
			{Name: "StartTime", Sort: true}, {Name: "FinishTime"},
			{Name: "Result", AutoColor: true},
		},
	}
	// trace
	pt.AddItems(action.Events)
	common.PrintPrettyTable(pt, long)

	for _, item := range action.Events {
		if item.Traceback == "" {
			continue
		}
		if long {
			fmt.Printf("Event %s tracback:\n", item.Event)
			println(item.Traceback)
		} else {
			console.Warn("use --long flags to show tracebacks")
		}
	}
}

var serverAction = &cobra.Command{
	Use:   "action <server> [request-id]",
	Short: "Get server action(s)",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		requestId := ""
		if len(args) == 2 {
			requestId = args[1]
		}
		client := common.DefaultClient()
		server, err := client.NovaV2().Server().Find(args[0])
		var serverId string
		if err == nil {
			serverId = server.Id
		} else {
			serverId = args[0]
		}
		if *serverActionFlags.Spend {
			listServerActionsWithSpend(serverId, *serverActionFlags.Name, requestId,
				*serverActionFlags.Last, *serverActionFlags.Long)
		} else if requestId == "" {
			listServerActions(serverId, *serverActionFlags.Name, *serverActionFlags.Last, *serverActionFlags.Long)
		} else {
			showAction(serverId, requestId, *serverActionFlags.Long)
		}
	},
}

func init() {
	serverActionFlags = flags.ServerActionFlags{
		Name:  serverAction.Flags().StringP("name", "n", "", "Filter by action name"),
		Spend: serverAction.Flags().Bool("spend", false, "List action events with spend"),
		Last:  serverAction.Flags().Int("last", 0, "Get last N actions"),

		Long: serverAction.Flags().BoolP("long", "l", false, "List additional fields in output"),
	}

	Server.AddCommand(serverAction)
}
