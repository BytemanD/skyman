package server_actions

import (
	"fmt"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/jedib0t/go-pretty/v6/text"
)

type ServerActionEvents struct {
	ServerId  string
	Action    string
	RequestId string
	Events    nova.InstanceActionEvents
}

func PrintServerEvents(client *openstack.Openstack) (string, error) {
	fmt.Println("Server events")
	return "", nil
	// serverEventReport := []ServerActionEvents{}
	// pt := common.PrettyTable{
	// 	Style: common.STYLE_LIGHT,
	// 	ShortColumns: []common.Column{
	// 		{Name: "ServerId"},
	// 		{Name: "Action"},
	// 		{Name: "RequestId"},
	// 		{Name: "Events", Slot: func(item interface{}) interface{} {
	// 			p := item.(ServerActionEvents)
	// 			eventResult := []string{}
	// 			for _, event := range p.Events {
	// 				eventResult = append(eventResult, fmt.Sprintf("%s(%s)", event.Event, event.Result))
	// 			}
	// 			return strings.Join(eventResult, "\n")
	// 		}},
	// 	},
	// }
	// for _, task := range TestTasks {
	// 	actions, err := client.NovaV2().Server().ListActionsWithEvents(task.ServerId, "", "", 0)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	for _, action := range actions {
	// 		serverEvents := ServerActionEvents{
	// 			ServerId:  task.ServerId,
	// 			RequestId: action.RequestId,
	// 			Action:    action.Action,
	// 			Events:    action.Events,
	// 		}
	// 		serverEventReport = append(serverEventReport, serverEvents)
	// 	}

	// }
	// pt.AddItems(serverEventReport)

	// fmt.Println("server events:")
	// return common.PrintPrettyTable(pt, false), nil
}

// action result
type ActionResult struct {
	Action string `json:"action"`
	Error  error  `json:"error"`
}

func (r *ActionResult) SetResult(err error) {
	r.Error = err
}

// Worker report
type WorkerReport struct {
	TestId  int            `json:"testId"`
	Server  string         `json:"server"`
	Error   error          `json:"error"`
	Results []ActionResult `json:"results"`
}

func (r *WorkerReport) Init(testId int, server string) {
	r.TestId = testId
	r.Server = server
}

func (r WorkerReport) HasError() bool {
	if r.Error != nil {
		return true
	}
	for _, report := range r.Results {
		if report.Error != nil {
			return true
		}
	}
	return false
}

func (t *WorkerReport) GetResultEmoji() string {
	if t.HasError() {
		return "😭"
	}
	return "😄"
}

type CaseReport struct {
	Name          string         `json:"name"`
	Workers       int            `json:"workers"`
	Actions       string         `json:"actions"`
	WorkerReports []WorkerReport `json:"reports"`
}

func (r *CaseReport) NameOrACtions() string {
	if r.Name != "" {
		return r.Name
	}
	return r.Actions
}

type ReportItem struct {
	NameOrActions string
	Workers       int
	Servers       string
	ResultEmojis  string
	Details       string
}

func PrintCaseReports(caseReports []CaseReport) {
	items := []ReportItem{}
	for _, caseReport := range caseReports {
		servers, details, emojis := []string{}, []string{}, []string{}
		for _, workerReport := range caseReport.WorkerReports {
			servers = append(servers, workerReport.Server)
			emojis = append(emojis, workerReport.GetResultEmoji())
			if workerReport.Error != nil {
				details = append(details, workerReport.Error.Error())
			}
		}
		item := ReportItem{
			NameOrActions: caseReport.NameOrACtions(),
			Workers:       caseReport.Workers,
			ResultEmojis:  strings.Join(emojis, "\n"),
			Servers:       strings.Join(servers, "\n"),
			Details:       strings.Join(details, "\n"),
		}
		items = append(items, item)
	}

	fmt.Println("Report:")
	pt := common.PrettyTable{
		Style:             common.STYLE_LIGHT,
		StyleSeparateRows: true,
		ShortColumns: []common.Column{
			{Name: "NameOrActions"},
			{Name: "Workers"},
			{Name: "Servers"},
			{Name: "ResultEmojis", Text: "Result", Align: text.AlignCenter},
			{Name: "Details"},
		},
	}
	pt.AddItems(items)
	common.PrintPrettyTable(pt, false)
}
