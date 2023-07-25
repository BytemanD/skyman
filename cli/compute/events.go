package compute

import (
	"fmt"
	"os"

	"github.com/BytemanD/stackcrud/cli"
	"github.com/spf13/cobra"
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
		dataTable := cli.DataListTable{
			ShortHeaders: []string{"Action", "RequestId", "StartTime", "Message"},
			LongHeaders:  []string{"ProjectId", "UserId"},
			HeaderLabel: map[string]string{
				"RequestId":  "Request Id",
				"Start Time": "Start Time",
				"Project Id": "ProjectId",
				"User Id":    "UserId",
			},
		}
		dataTable.AddItems(actions)
		dataTable.Print(long)
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
		dataTable := cli.DataListTable{
			Title:        fmt.Sprintf("Action: %s", action.Action),
			ShortHeaders: []string{"Event", "StartTime", "FinishTime", "Result"},
			LongHeaders:  []string{"Host"},
			HeaderLabel: map[string]string{
				"StartTime":  "Start Time",
				"FinishTime": "Start Time",
			},
		}
		// trace
		tracbackMap := map[string]string{}
		for _, item := range action.Events {
			if item.Traceback != "" {
				tracbackMap[item.Event] = item.Traceback
			}
		}
		dataTable.AddItems(action.Events)
		dataTable.Print(long)
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
