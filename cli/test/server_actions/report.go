package server_actions

import (
	"fmt"

	"github.com/BytemanD/skyman/common"
	"github.com/jedib0t/go-pretty/v6/text"
)

type ServerReport struct {
	ServerId        string
	Result          string
	TotalActionsNum int
	FailedActinos   string
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
			{Name: "Result"},
			{Name: "TotalActionsNum", Align: text.AlignRight},
			{Name: "FailedActinos"},
		},
	}
	pt.AddItems(reportItems)
	return common.PrintPrettyTable(pt, false)
}
