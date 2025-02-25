package test

import (
	"strings"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/server_actions"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliActions *server_actions.ActionCountList

var TestServerAction = &cobra.Command{
	Use:   "server-actions <TASK FILE> [server]",
	Short: "Test server actions",
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return err
		}
		actions, _ := cmd.Flags().GetString("actions")
		if actions != "" {
			if testActions, err := server_actions.NewActionCountList(actions); err != nil {
				return err
			} else {
				cliActions = testActions
			}
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		testConf, err := common.LoadTaskConfig(args[0])
		if err != nil {
			console.Fatal("load task file %s failed: %s", args[0], err)
		}

		servers, _ := cmd.Flags().GetString("servers")
		// reportEvents, _ := cmd.Flags().GetBool("report-events")
		web, _ := cmd.Flags().GetBool("web")

		userServers := []string{}
		if servers != "" {
			userServers = strings.Split(servers, ",")
		}

		testCases := []server_actions.Case{}
		// 初始化用例
		client := common.DefaultClient()
		console.Info("test on region: %s", client.Region())
		if cliActions != nil {
			worker, _ := cmd.Flags().GetInt("worker")
			actionInterval, _ := cmd.Flags().GetInt("action-interval")
			testCase := server_actions.Case{
				Actions:    *cliActions,
				UseServers: userServers,
				Client:     client,
				Config: common.NewActionCaseConfig(common.CaseConfig{
					Workers:        worker,
					ActionInterval: actionInterval,
				}, testConf.Default),
			}
			testCases = append(testCases, testCase)
		} else {
			console.Info("Found %d case(s)", len(testConf.Cases))
			for _, actionCase := range testConf.Cases {
				if actionCase.Skip {
					console.Warn("skip case '%s'", lo.CoalesceOrEmpty(actionCase.Name, actionCase.Actions))
					continue
				}
				acl, err := server_actions.NewActionCountList(actionCase.Actions)
				if err != nil {
					console.Fatal("parse actions '%s' failed, %s", actionCase.Actions, err)
				}
				testCase := server_actions.Case{
					Name:       actionCase.Name,
					UseServers: userServers,
					Actions:    *acl,
					Client:     client,
					Config:     common.NewActionCaseConfig(actionCase.Config, testConf.Default),
				}
				testCases = append(testCases, testCase)
			}
		}
		console.Info("Execute %d case(s)", len(testCases))

		// 测试前检查
		// preTest(client)

		if web {
			go server_actions.RunSimpleWebServer(testConf.Web)
		}

		caseReports := []server_actions.CaseReport{}
		for _, testCase := range testCases {
			testCase.Start()
			caseReports = append(caseReports, testCase.Report())
		}
		server_actions.PrintCaseReports(caseReports)
		// if reportEvents {
		// 	testCase.PrintServerEvents()
		// }
		if web {
			server_actions.WaitWebServer()
		}
	},
}

func init() {
	supportedActions := []string{}

	for _, actions := range arrayutils.SplitStrings(server_actions.ValidActions().Keys(), 5) {
		supportedActions = append(supportedActions, strings.Join(actions, ", "))
	}
	TestServerAction.Flags().StringP("actions", "A", "", "Test actions\nFormat: <action>[:count], "+
		"multiple actions separate by ','.\nExample: reboot,live_migrate:3\n"+
		"Actions: "+strings.Join(supportedActions, ",\n         "),
	)

	TestServerAction.Flags().Int("action-interval", 0, "Action interval")
	TestServerAction.Flags().Int("total", 0, i18n.T("theNumOfTask"))
	TestServerAction.Flags().Int("worker", 0, i18n.T("theNumOfWorker"))
	TestServerAction.Flags().String("servers", "", "Use existing servers")
	// TestServerAction.Flags().Bool("report-events", false, i18n.T("reportServerEvents"))
	TestServerAction.Flags().Bool("web", false, "Start web server")

	viper.BindPFlag("test.total", TestServerAction.Flags().Lookup("total"))
	viper.BindPFlag("test.workers", TestServerAction.Flags().Lookup("worker"))
	viper.BindPFlag("test.actionInterval", TestServerAction.Flags().Lookup("action-interval"))

}
