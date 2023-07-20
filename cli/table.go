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
}

func (dataTable DataListTable) AddItems(items []interface{}) {
	dataTable.Items = append(dataTable.Items, items...)
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

	for _, data := range dataTable.Items {
		reflectValue := reflect.ValueOf(data)
		row := table.Row{}
		for _, name := range titles {
			value := reflectValue.FieldByName(name)
			row = append(row, value)
		}
		tableWriter.AppendRow(row)
	}
	tableWriter.SortBy(dataTable.SortBy)
	tableWriter.SetColumnConfigs(dataTable.ColumnConfigs)
	tableWriter.Render()
}
