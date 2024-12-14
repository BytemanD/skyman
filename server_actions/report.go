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
		actions, err := client.NovaV2().Server().ListActionsWithEvents(task.ServerId, "", "", 0)
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
	Id             int      `json:"id"`
	ServerId       string   `json:"serverId"`
	TotalActions   []string `json:"totalActions"`
	SuccessActions []string `json:"successActions"`
	SkipActions    []string `json:"skipActions"`
	FailedActions  []string `json:"failedActions"`
	Stage          string   `json:"stage"`
	Result         string   `json:"result"`
	Message        string   `json:"message"`
	Error          error    `json:"error"`
	Complated      int      `json:"completed"`
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
func (t *TestTask) MarkSuccess() {
	t.SetStage("")
	t.setResult("success", "")
}
func (t *TestTask) MarkFailed(message string, err error) {
	t.SetStage("")
	t.setResult("failed", message)
	t.Error = err
}
func (t *TestTask) MarkWarning() {
	t.SetStage("")
	t.setResult("warning", t.Message)
}
func (t *TestTask) GetResultEmoji() string {
	switch t.Result {
	case "success":
		return "😄"
	case "warning":
		return "😥"
	case "failed":
		return "😭"
	default:
		return "😶"
	}
}

func (t *TestTask) AddFailedAction(action string) {
	t.FailedActions = append(t.FailedActions, action)
}
func (t TestTask) GetError() error {
	if t.Result == "failed" {
		return fmt.Errorf(t.Message)
	}
	return nil
}
func (t TestTask) AllSuccess() bool {
	return len(t.TotalActions) == len(t.SuccessActions)
}
func (t TestTask) HasFailed() bool {
	return len(t.FailedActions) > 0
}
func (t TestTask) HasSkip() bool {
	return len(t.SkipActions) > 0
}
func (t TestTask) GetResultString() string {
	return fmt.Sprintf("all actions: %d, success: %d, failed: %d, skip: %d",
		len(t.TotalActions), len(t.SuccessActions), len(t.FailedActions),
		len(t.SkipActions))
}

var TestTasks = []*TestTask{}

func PrintTestTasks(reports []TestTask) {
	fmt.Println("result:")
	pt := common.PrettyTable{
		Style: common.STYLE_LIGHT,
		ShortColumns: []common.Column{
			{Name: "#", Align: text.AlignCenter, Slot: func(item interface{}) interface{} {
				p := item.(TestTask)
				return p.GetResultEmoji()
			}},
			{Name: "ServerId"},
			{Name: "Actions", Slot: func(item interface{}) interface{} {
				p := item.(TestTask)
				return strings.Join(p.TotalActions, ",")
			}},
			{Name: "Result", Text: "Success/Skip/Failed", Align: text.AlignCenter,
				Slot: func(item interface{}) interface{} {
					p, _ := item.(TestTask)
					return fmt.Sprintf("%d/%d/%d", len(p.SuccessActions),
						len(p.SkipActions), len(p.FailedActions),
					)
				}},
			{Name: "Message", Slot: func(item interface{}) interface{} {
				p, _ := item.(TestTask)
				if p.Message != "" {
					return p.Message
				} else if p.HasFailed() {
					return fmt.Sprintf("failed actions: %s", strings.Join(p.FailedActions, ", "))
				} else {
					return ""
				}
			}},
		},
	}
	pt.AddItems(reports)
	common.PrintPrettyTable(pt, false)
}
