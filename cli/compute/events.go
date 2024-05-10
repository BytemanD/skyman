package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
)

func listServerActions(server string, actionName string, last int, long bool) {
	client := openstack.DefaultClient()
	actions, err := client.NovaV2().Servers().ListActions(server)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "Action"}, {Name: "RequestId"}, {Name: "StartTime", Sort: true},
			{Name: "Message"},
		},
		LongColumns: []common.Column{
			{Name: "ProjectId"}, {Name: "UserId"},
		},
	}
	if last == 0 {
		last = len(actions)
	}
	for _, action := range actions {
		if actionName != "" && action.Action != actionName {
			continue
		}
		pt.Items = append(pt.Items, action)
	}
	pt.Items = common.LastItems(pt.Items, last)
	common.PrintPrettyTable(pt, long)
}
func listServerActionsWithSpend(server string, actionName string, requestId string, last int, long bool) {
	client := openstack.DefaultClient()
	actionsWithEvents, err := client.NovaV2().Servers().ListActionsWithEvents(
		server, actionName, requestId, last)
	utility.LogError(err, "get server actions and events failed", true)

	pt := common.PrettyTable{
		ShortColumns: []common.Column{
			{Name: "Name", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.InstanceActionAndEvents)
				return p.InstanceAction.Action
			}},
			{Name: "RequestId", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.InstanceActionAndEvents)
				return p.InstanceAction.RequestId
			}},
			{Name: "StartTime", Slot: func(item interface{}) interface{} {
				p, _ := item.(nova.InstanceActionAndEvents)
				return p.InstanceAction.StartTime
			}},
			{Name: "SpendTime", Text: "Spend time (seconds)",
				Slot: func(item interface{}) interface{} {
					p, _ := (item).(nova.InstanceActionAndEvents)
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
						p, _ := (item).(nova.InstanceActionAndEvents)
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
func showAction(server string, requestId string, long bool) {
	client := openstack.DefaultClient()
	action, err := client.NovaV2().Servers().ShowAction(server, requestId)
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
			fmt.Println(item.Traceback)
		} else {
			logging.Warning("use --long flags to show tracebacks")
		}
	}
}

var serverAction = &cobra.Command{
	Use:   "action <server> [request-id]",
	Short: "Get server action(s)",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		server, requestId := args[0], ""
		long, _ := cmd.Flags().GetBool("long")
		actionName, _ := cmd.Flags().GetString("name")
		spend, _ := cmd.Flags().GetBool("spend")
		last, _ := cmd.Flags().GetInt("last")

		if len(args) == 2 {
			requestId = args[1]
		}

		if spend {
			listServerActionsWithSpend(server, actionName, requestId, last, long)
		} else if requestId == "" {
			listServerActions(server, actionName, last, long)
		} else {
			showAction(server, requestId, long)
		}
	},
}

func init() {
	serverAction.Flags().BoolP("long", "l", false, "List additional fields in output")

	serverAction.Flags().Bool("spend", false, "List action events with spend")
	serverAction.Flags().StringP("name", "n", "", "Filter by action name")
	serverAction.Flags().StringP("request-id", "r", "", "Filter by request id")
	serverAction.Flags().Int("last", 0, "Get last N actions")

	// serverAction.AddCommand(actionList, actionShow, actionSpend)

	Server.AddCommand(serverAction)
}
