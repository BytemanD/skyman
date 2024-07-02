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

type ServerReport struct {
	ServerId        string `json:"serverId"`
	Result          string `json:"result"`
	TotalActionsNum int    `json:"totalActionNum"`
	FailedActinos   string `json:"failedActions"`
}

func (r *ServerReport) AddFailedAction(action string) {
	if r.FailedActinos == "" {
		r.FailedActinos = action
	} else {
		r.FailedActinos = fmt.Sprintf("%s,%s", r.FailedActinos, action)
	}
}

func PrintReport(reportItems []ServerReport) string {
	pt := common.PrettyTable{
		Style: common.STYLE_LIGHT,
		ShortColumns: []common.Column{
			{Name: "ServerId"},
			{Name: "Result", Slot: func(item interface{}) interface{} {
				p := item.(ServerReport)
				return utility.ColorString(p.Result)
			}},
			{Name: "TotalActionsNum", Align: text.AlignRight},
			{Name: "FailedActinos"},
		},
	}
	pt.AddItems(reportItems)
	return common.PrintPrettyTable(pt, false)
}

type ServerActionEvents struct {
	ServerId  string
	Action    string
	RequestId string
	Events    nova.InstanceActionEvents
}

func PrintServerEvents(client *openstack.Openstack, serverIds []string) (string, error) {
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
	for _, serverId := range serverIds {
		actions, err := client.NovaV2().Servers().ListActionsWithEvents(serverId, "", "", 0)
		if err != nil {
			return "", err
		}
		for _, action := range actions {
			serverEvents := ServerActionEvents{
				ServerId:  serverId,
				RequestId: action.RequestId,
				Action:    action.Action,
				Events:    action.Events,
			}
			serverEventReport = append(serverEventReport, serverEvents)
		}

	}
	pt.AddItems(serverEventReport)
	return common.PrintPrettyTable(pt, false), nil
}
