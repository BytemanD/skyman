package test

import (
	"os"
	"strings"

	"github.com/BytemanD/easygo/pkg/arrayutils"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/common/i18n"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/server_actions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

//	func preTest(client *openstack.Openstack) {
//		logging.Info("check flavors ...")
//		for _, flavorId := range common.TASK_CONF.Flavors {
//			flavor, err := client.NovaV2().Flavor().Found(flavorId, false)
//			utility.LogError(err, fmt.Sprintf("get flavor %s failed", flavorId), true)
//			server_actions.TEST_FLAVORS = append(server_actions.TEST_FLAVORS, *flavor)
//		}
//		logging.Info("check images ...")
//		for _, idOrName := range common.TASK_CONF.Images {
//			_, err := client.GlanceV2().Images().Found(idOrName)
//			utility.LogError(err, fmt.Sprintf("get image %s failed", idOrName), true)
//		}
//		logging.Info("check networks ...")
//		for _, idOrName := range common.TASK_CONF.Networks {
//			_, err := client.NeutronV2().Network().Show(idOrName)
//			utility.LogError(err, fmt.Sprintf("get network %s failed", idOrName), true)
//		}
//	}

var cliActions *server_actions.ActionCountList

func actionCaseConfig(config common.CaseConfig, def common.CaseConfig) common.CaseConfig {
	caseConfig := config
	if len(config.Flavors) == 0 {
		caseConfig.Flavors = def.Flavors
	}
	if len(config.Images) == 0 {
		caseConfig.Images = def.Images
	}
	if config.Workers <= 0 {
		caseConfig.Workers = max(def.Workers, 1)
	}
	if config.ActionInterval <= 0 {
		caseConfig.ActionInterval = max(def.ActionInterval, 0)
	}
	return caseConfig
}

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
		if err := common.LoadTaskConfig(args[0]); err != nil {
			logging.Error("load task file %s failed: %s", args[0], err)
			os.Exit(1)
		}

		servers, _ := cmd.Flags().GetString("servers")
		reportEvents, _ := cmd.Flags().GetBool("report-events")
		web, _ := cmd.Flags().GetBool("web")

		userServers := []string{}
		if servers != "" {
			userServers = strings.Split(servers, ",")
		}

		testCases := []server_actions.Case{}
		// 初始化用例
		client := openstack.DefaultClient()
		if cliActions != nil {
			worker, _ := cmd.Flags().GetInt("worker")
			actionInterval, _ := cmd.Flags().GetInt("action-interval")
			testCase := server_actions.Case{
				Actions:    *cliActions,
				UseServers: userServers,
				Client:     client,
				Config: actionCaseConfig(common.CaseConfig{
					Workers:        worker,
					ActionInterval: actionInterval,
				}, common.TASK_CONF.Default),
			}
			testCases = append(testCases, testCase)
		} else {
			for _, actionCase := range common.TASK_CONF.Cases {
				acl, err := server_actions.NewActionCountList(actionCase.Actions)
				if err != nil {
					logging.Fatal("parse actions failed: %s", actionCase.Actions)
				}
				testCase := server_actions.Case{
					Actions: *acl,
					Client:  client,
					Config:  actionCaseConfig(actionCase.Config, common.TASK_CONF.Default),
				}

				testCases = append(testCases, testCase)
			}
		}

		logging.Info("Found %d case(s)", len(testCases))

		// 测试前检查
		// preTest(client)

		if web {
			go server_actions.RunSimpleWebServer()
		}

		for _, testCase := range testCases {
			err := testCase.Start()
			if err == nil {
				testCase.PrintReport()
				if reportEvents {
					testCase.PrintServerEvents()
				}
			}
		}
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
	TestServerAction.Flags().Bool("report-events", false, i18n.T("reportServerEvents"))
	TestServerAction.Flags().Bool("web", false, "Start web server")

	viper.BindPFlag("test.total", TestServerAction.Flags().Lookup("total"))
	viper.BindPFlag("test.workers", TestServerAction.Flags().Lookup("worker"))
	viper.BindPFlag("test.actionInterval", TestServerAction.Flags().Lookup("action-interval"))

}
