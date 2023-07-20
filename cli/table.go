package cli

import (
	"os"
	"reflect"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// TODO: move to easygo
type DataListTable struct {
	ShortHeaders  []string
	LongHeaders   []string
	HeaderLabel   map[string]string
	Items         []interface{}
	SortBy        []table.SortBy
	ColumnConfigs []table.ColumnConfig
	Slots         map[string]func(item interface{}) interface{}
}

func (dataTable DataListTable) Print(long bool) {
	tableWriter := table.NewWriter()
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)

	headerRow := table.Row{}
	titles := dataTable.ShortHeaders
	if long {
		titles = append(titles, dataTable.LongHeaders...)
	}
	for _, header := range titles {
		var title string
		if _, ok := dataTable.HeaderLabel[header]; ok {
			title = dataTable.HeaderLabel[header]
		} else {
			title = header
		}
		headerRow = append(headerRow, title)
	}
	tableWriter.AppendHeader(headerRow)

	for _, item := range dataTable.Items {
		reflectValue := reflect.ValueOf(item)
		row := table.Row{}
		for _, name := range titles {
			var value interface{}
			if _, ok := dataTable.Slots[name]; ok {
				value = dataTable.Slots[name](item)
			} else {
				value = reflectValue.FieldByName(name)
			}
			row = append(row, value)
		}
		tableWriter.AppendRow(row)
	}
	tableWriter.SortBy(dataTable.SortBy)
	tableWriter.SetColumnConfigs(dataTable.ColumnConfigs)
	tableWriter.Render()
}
