package common

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Column struct {
	Name string
	Text string
	// 只有 Table.Style 等于 light 是才会生效
	AutoColor bool
	Slot      func(item interface{}) interface{}
}

type PrettyTable struct {
	Title             string
	ShortColumns      []Column
	LongColumns       []Column
	Items             []interface{}
	SortBy            []table.SortBy
	ColumnConfigs     []table.ColumnConfig
	Style             string
	StyleSeparateRows bool
	tableWriter       table.Writer
}

func (pt *PrettyTable) AddItems(items interface{}) {
	value := reflect.ValueOf(items)
	for i := 0; i < value.Len(); i++ {
		pt.Items = append(pt.Items, value.Index(i).Interface())
	}
}
func (pt *PrettyTable) CleanItems() {
	if len(pt.Items) > 0 {
		pt.Items = []interface{}{}
	}
}
func (pt *PrettyTable) SetStyleLight() {
	pt.Style = STYLE_LIGHT
}

func (pt *PrettyTable) getTableWriter() table.Writer {
	if pt.tableWriter == nil {
		pt.tableWriter = table.NewWriter()
		if pt.Style == STYLE_LIGHT {
			pt.tableWriter.SetStyle(table.StyleLight)
			pt.tableWriter.Style().Color.Header = text.Colors{text.FgBlue, text.Bold}
			pt.tableWriter.Style().Color.Border = text.Colors{text.FgBlue}
			pt.tableWriter.Style().Color.Separator = text.Colors{text.FgBlue}
		}
		pt.tableWriter.Style().Format.Header = text.FormatDefault
		pt.tableWriter.Style().Options.SeparateRows = pt.StyleSeparateRows
		pt.tableWriter.SortBy(pt.SortBy)
		pt.tableWriter.SetColumnConfigs(pt.ColumnConfigs)
		pt.tableWriter.SetOutputMirror(os.Stdout)
	}
	return pt.tableWriter
}
func (pt *PrettyTable) ReInit() {
	pt.tableWriter = nil
}

func (pt PrettyTable) Print(long bool) {
	tableWriter := pt.getTableWriter()

	headerRow := table.Row{}
	columns := pt.ShortColumns
	if long {
		columns = append(columns, pt.LongColumns...)
	}
	for _, column := range columns {
		var title string
		if column.Text == "" {
			title = splitTitle(column.Name)
		} else {
			title = column.Text
		}
		headerRow = append(headerRow, title)
	}
	tableWriter.AppendHeader(headerRow)

	for _, item := range pt.Items {
		reflectValue := reflect.ValueOf(item)
		row := table.Row{}
		for _, column := range columns {
			var value interface{}
			if column.Slot != nil {
				value = column.Slot(item)
			} else {
				value = reflectValue.FieldByName(column.Name)
			}
			if column.AutoColor && pt.Style == STYLE_LIGHT {
				value = pt.FormatString(fmt.Sprint(value))
			}
			row = append(row, value)
		}
		tableWriter.AppendRow(row)
	}
	if pt.Title != "" {
		tableWriter.SetTitle(pt.Title)
		tableWriter.Style().Title.Align = text.AlignCenter
	}

	tableWriter.Render()
	fmt.Printf("Total items: %d\n", len(pt.Items))
}

func (pt PrettyTable) FormatString(s string) string {
	if pt.Style == STYLE_LIGHT {
		return BaseColorFormatter.Format(s)
	} else {
		return s
	}
}

func (pt PrettyTable) PrintJson() {
	output, err := GetIndentJson(pt.Items)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}

func (pt PrettyTable) PrintYaml() {
	output, err := GetYaml(pt.Items)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}
