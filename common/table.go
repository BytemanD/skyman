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
	// 废弃，使用 Field.Slot
	Slots map[string]func(item interface{}) interface{}
	Title string
	Style string
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

	headerRow := table.Row{"Field", "Value"}
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
		} else if _, ok := dataTable.Slots[field.Name]; ok {
			fieldValue = dataTable.Slots[field.Name](dataTable.Item)
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

type DataListTable struct {
	ShortHeaders      []string
	LongHeaders       []string
	HeaderLabel       map[string]string
	Items             []interface{}
	SortBy            []table.SortBy
	AutoFormat        []string
	ColumnConfigs     []table.ColumnConfig
	Slots             map[string]func(item interface{}) interface{}
	Title             string
	StyleSeparateRows bool
	Style             string
}

func (dataTable *DataListTable) AddItems(items interface{}) {
	value := reflect.ValueOf(items)
	for i := 0; i < value.Len(); i++ {
		dataTable.Items = append(dataTable.Items, value.Index(i).Interface())
	}
}
func (dataTable *DataListTable) CleanItems() {
	if len(dataTable.Items) > 0 {
		dataTable.Items = []interface{}{}
	}
}
func (dataTable *DataListTable) SetStyleLight() {
	dataTable.Style = STYLE_LIGHT
}
func (dataTable DataListTable) Print(long bool) {
	tableWriter := table.NewWriter()
	if dataTable.Style == STYLE_LIGHT {
		tableWriter.SetStyle(table.StyleLight)
		tableWriter.Style().Color.Header = text.Colors{text.FgBlue, text.Bold}
		tableWriter.Style().Color.Border = text.Colors{text.FgBlue}
		tableWriter.Style().Color.Separator = text.Colors{text.FgBlue}
	}

	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.Style().Options.SeparateRows = dataTable.StyleSeparateRows
	tableWriter.SortBy(dataTable.SortBy)
	tableWriter.SetColumnConfigs(dataTable.ColumnConfigs)

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
			title = splitTitle(header)
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
			for _, s := range dataTable.AutoFormat {
				if s == name {
					value = dataTable.FormatString(fmt.Sprint(value))
				}
			}
			row = append(row, value)
		}
		tableWriter.AppendRow(row)
	}
	if dataTable.Title != "" {
		tableWriter.SetTitle(dataTable.Title)
		tableWriter.Style().Title.Align = text.AlignCenter
	}

	tableWriter.Render()
	fmt.Printf("Total items: %d\n", len(dataTable.Items))
}

func (dataTable DataListTable) PrintJson() {
	output, err := GetIndentJson(dataTable.Items)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}

func (dataTable DataListTable) PrintYaml() {
	output, err := GetYaml(dataTable.Items)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}

func (dataTable DataListTable) FormatString(s string) string {
	if dataTable.Style == STYLE_LIGHT {
		return BaseColorFormatter.Format(s)
	} else {
		return s
	}
}
