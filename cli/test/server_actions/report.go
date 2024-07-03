package server_actions

import (
	"fmt"
	"strings"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack"
	"github.com/BytemanD/skyman/openstack/model/nova"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/text"
)

type ServerActionEvents struct {
	ServerId  string
	Action    string
	RequestId string
	Events    nova.InstanceActionEvents
}

func PrintServerEvents(client *openstack.Openstack) (string, error) {
	serverEventReport := []ServerActionEvents{}
	pt := common.PrettyTable{
		Style: common.STYLE_LIGHT,
		ShortColumns: []common.Column{
			{Name: "ServerId"},
			{Name: "Action"},
			{Name: "RequestId"},
			{Name: "Events", Slot: func(item interface{}) interface{} {
				p := item.(ServerActionEvents)
				eventResult := []string{}
				for _, event := range p.Events {
					eventResult = append(eventResult, fmt.Sprintf("%s(%s)", event.Event, event.Result))
				}
				return strings.Join(eventResult, "\n")
			}},
		},
	}
	for _, task := range TestTasks {
		actions, err := client.NovaV2().Servers().ListActionsWithEvents(task.ServerId, "", "", 0)
		if err != nil {
			return "", err
		}
		for _, action := range actions {
			serverEvents := ServerActionEvents{
				ServerId:  task.ServerId,
				RequestId: action.RequestId,
				Action:    action.Action,
				Events:    action.Events,
			}
			serverEventReport = append(serverEventReport, serverEvents)
		}

	}
	pt.AddItems(serverEventReport)

	fmt.Println("server events:")
	return common.PrintPrettyTable(pt, false), nil
}

type TestTask struct {
	Id            int      `json:"id"`
	ServerId      string   `json:"serverId"`
	Total         int      `json:"total"`
	Complated     int      `json:"completed"`
	Stage         string   `json:"stage"`
	Result        string   `json:"result"`
	Message       string   `json:"message"`
	FailedActions []string `json:"failedActions"`
}

func (t *TestTask) SetStage(stage string) {
	t.Stage = stage
}
func (t *TestTask) setResult(result string, message string) {
	t.Result = result
	t.Message = message
}
func (t *TestTask) IncrementCompleted() {
	t.Complated += 1
}
func (t *TestTask) Success() {
	t.SetStage("")
	t.setResult("success", "")
}
func (t *TestTask) Failed(message string) {
	t.SetStage("")
	t.setResult("failed", message)
}
func (t *TestTask) AddFailedAction(action string) {
	t.FailedActions = append(t.FailedActions, action)
}
func (t TestTask) GetError() error {
	if t.Result == "failedd" {
		return fmt.Errorf(t.Message)
	}
	return nil
}

var TestTasks = []*TestTask{}

func PrintTestTasks() string {
	fmt.Println("result:")
	pt := common.PrettyTable{
		Style: common.STYLE_LIGHT,
		ShortColumns: []common.Column{
			{Name: "ServerId"},
			{Name: "Result", Slot: func(item interface{}) interface{} {
				p := item.(TestTask)
				return utility.ColorString(p.Result)
			}},
			{Name: "Total", Align: text.AlignRight},
			{Name: "FailedActinos", Slot: func(item interface{}) interface{} {
				p := item.(TestTask)
				return strings.Join(p.FailedActions, ", ")
			}},
		},
	}
	pt.AddItems(TestTasks)
	return common.PrintPrettyTable(pt, false)
}
