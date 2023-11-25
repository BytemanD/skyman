package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/compute"
)

var serverAction = &cobra.Command{Use: "action"}

var actionList = &cobra.Command{
	Use:   "list <server>",
	Short: "List server actions",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		long, _ := cmd.Flags().GetBool("long")
		actions, err := client.ComputeClient().ServerActionList(args[0])
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
		pt.AddItems(actions)
		common.PrintPrettyTable(pt, long)
	},
}

var actionShow = &cobra.Command{
	Use:   "show <server> <request id>",
	Short: "Show server action",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		long, _ := cmd.Flags().GetBool("long")
		id := args[0]
		requestId := args[1]
		action, err := client.ComputeClient().ServerActionShow(id, requestId)
		common.LogError(err, "get server action failed", true)
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

		if long {
			for _, item := range action.Events {
				if item.Traceback == "" {
					continue
				}
				fmt.Printf("Event %s tracback:\n", item.Event)
				fmt.Println(item.Traceback)
			}
		}
	},
}

var actionSpend = &cobra.Command{
	Use:   "spend <server>",
	Short: "Get server actions spend time",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		long, _ := cmd.Flags().GetBool("long")
		actionName, _ := cmd.Flags().GetString("name")
		requestId, _ := cmd.Flags().GetString("request-id")

		actionsWithEvents, err := client.ComputeClient().ServerActionsWithEvents(args[0], actionName, requestId)
		common.LogError(err, "get server actions and events failed", true)

		pt := common.PrettyTable{
			ShortColumns: []common.Column{
				{Name: "Name", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.InstanceActionAndEvents)
					return p.InstanceAction.Action
				}},
				{Name: "RequestId", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.InstanceActionAndEvents)
					return p.InstanceAction.RequestId
				}},
				{Name: "SpendTime", Slot: func(item interface{}) interface{} {
					p, _ := (item).(compute.InstanceActionAndEvents)
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
				if columnIndex < 0 {
					pt.LongColumns = append(pt.LongColumns,
						common.Column{
							Name: event.Event,
							SlotColumn: func(item interface{}, column common.Column) interface{} {
								p, _ := (item).(compute.InstanceActionAndEvents)
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
		}
		pt.AddItems(actionsWithEvents)
		common.PrintPrettyTable(pt, long)
	},
}

func init() {
	serverAction.PersistentFlags().BoolP("long", "l", false, "List additional fields in output")

	actionSpend.Flags().StringP("name", "n", "", "Filter by action name")
	actionSpend.Flags().StringP("request-id", "r", "", "Filter by request id")

	serverAction.AddCommand(actionList, actionShow, actionSpend)

	Server.AddCommand(serverAction)
}
