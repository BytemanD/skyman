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
			} else {
				details = append(details, "")
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
