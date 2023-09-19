package compute

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/BytemanD/skyman/cli"
	"github.com/BytemanD/skyman/common"
)

var serverAction = &cobra.Command{Use: "action"}

var actionList = &cobra.Command{
	Use:   "list <server>",
	Short: "List server actions",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		client := cli.GetClient()
		long, _ := cmd.Flags().GetBool("long")
		actions, err := client.Compute.ServerActionList(args[0])
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
		action, err := client.Compute.ServerActionShow(id, requestId)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		pt := common.PrettyTable{
			Title: fmt.Sprintf("Action: %s", action.Action),
			ShortColumns: []common.Column{
				{Name: "Event"}, {Name: "Host"},
				{Name: "StartTime", Sort: true}, {Name: "FinishTime"},
				{Name: "Result", AutoColor: true},
			},
			LongColumns: []common.Column{
				{Name: "ProjectId"}, {Name: "UserId"},
			},
		}
		// trace
		tracbackMap := map[string]string{}
		for _, item := range action.Events {
			if item.Traceback != "" {
				tracbackMap[item.Event] = item.Traceback
			}
		}
		pt.AddItems(action.Events)
		common.PrintPrettyTable(pt, long)
		if long {
			for k, v := range tracbackMap {
				fmt.Printf("Event %s tracback:\n", k)
				fmt.Println(v)
			}
		}
	},
}

func init() {
	serverAction.PersistentFlags().BoolP("long", "l", false, "List additional fields in output")

	serverAction.AddCommand(actionList)
	serverAction.AddCommand(actionShow)

	Server.AddCommand(serverAction)
}
