package common

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// TODO: use easygo

var (
	STYLE_LIGHT = "light"
)

type Field struct {
	Name string
	Text string
	Slot func(item interface{}) interface{}
}

type DataTable struct {
	ShortFields []Field
	LongFields  []Field
	Item        interface{}
	Title       string
	Style       string
}

func (dataTable DataTable) Print(long bool) {
	tableWriter := table.NewWriter()
	if dataTable.Style == STYLE_LIGHT {
		tableWriter.SetStyle(table.StyleLight)
		tableWriter.Style().Color.Header = text.Colors{text.FgBlue, text.Bold}
		tableWriter.Style().Color.Border = text.Colors{text.FgBlue}
		tableWriter.Style().Color.Separator = text.Colors{text.FgBlue}
	}

	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)

	headerRow := table.Row{"Property", "Value"}
	fields := dataTable.ShortFields
	if long {
		fields = append(fields, dataTable.LongFields...)
	}
	tableWriter.AppendHeader(headerRow)
	reflectValue := reflect.ValueOf(dataTable.Item)
	for _, field := range fields {
		var (
			fieldValue interface{}
			fieldLabel string
		)
		if field.Text == "" {
			fieldLabel = splitTitle(field.Name)
		} else {
			fieldLabel = field.Text
		}
		if field.Slot != nil {
			fieldValue = field.Slot(dataTable.Item)
		} else {
			fieldValue = reflectValue.FieldByName(field.Name)
		}
		tableWriter.AppendRow(table.Row{fieldLabel, fieldValue})
	}
	if dataTable.Title != "" {
		tableWriter.SetTitle(dataTable.Title)
		tableWriter.Style().Title.Align = text.AlignCenter
	}
	tableWriter.Render()
}
func (dt DataTable) PrintJson() {
	output, err := GetIndentJson(dt.Item)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}

func (dt DataTable) PrintYaml() {
	output, err := GetYaml(dt.Item)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}
