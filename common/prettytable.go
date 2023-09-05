package common

import (
	"fmt"
	"os"
	"reflect"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func splitTitle(s string) string {
	newStr := ""
	for _, c := range s {
		if c < 91 && newStr != "" {
			newStr += " " + string(c)
		} else {
			newStr += string(c)
		}
	}
	return newStr
}

var (
	STYLE_LIGHT = "light"
)

type Column struct {
	Name string
	Text string
	// 只有 Table.Style 等于 light 是才会生效
	AutoColor  bool
	ForceColor bool
	Slot       func(item interface{}) interface{}
	Sort       bool
	SortMode   table.SortMode
}

type PrettyTable struct {
	Title             string
	ShortColumns      []Column
	LongColumns       []Column
	Items             []interface{}
	ColumnConfigs     []table.ColumnConfig
	Style             string
	StyleSeparateRows bool
	HideTotalItems    bool
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

		pt.tableWriter.SetColumnConfigs(pt.ColumnConfigs)
		pt.tableWriter.SetOutputMirror(os.Stdout)
	}
	return pt.tableWriter
}
func (pt *PrettyTable) ReInit() {
	pt.tableWriter = nil
}
func (pt PrettyTable) getSortName(column Column) string {
	if column.Text != "" {
		return column.Text
	} else {
		return column.Name
	}
}
func (pt PrettyTable) Print(long bool) {
	tableWriter := pt.getTableWriter()

	headerRow := table.Row{}
	columns := pt.ShortColumns
	if long {
		columns = append(columns, pt.LongColumns...)
	}
	sortBy := []table.SortBy{}
	for _, column := range columns {
		var title string
		if column.Text == "" {
			title = splitTitle(column.Name)
		} else {
			title = column.Text
		}
		if column.Sort {
			sortBy = append(sortBy,
				table.SortBy{Name: pt.getSortName(column), Mode: column.SortMode})
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
			if column.ForceColor || (column.AutoColor && pt.Style == STYLE_LIGHT) {
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
	// TODO: 当前只能按Columns 顺序排序
	tableWriter.SortBy(sortBy)
	tableWriter.Render()
	if !pt.HideTotalItems {
		fmt.Printf("Total items: %d\n", len(pt.Items))
	}
}

func (pt PrettyTable) FormatString(s string) string {
	return BaseColorFormatter.Format(s)
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

type PrettyItemTable struct {
	ShortFields []Column
	LongFields  []Column
	Item        interface{}
	Title       string
	Style       string
}

func (pt PrettyItemTable) Print(long bool) {
	tableWriter := table.NewWriter()
	if pt.Style == STYLE_LIGHT {
		tableWriter.SetStyle(table.StyleLight)
		tableWriter.Style().Color.Header = text.Colors{text.FgBlue, text.Bold}
		tableWriter.Style().Color.Border = text.Colors{text.FgBlue}
		tableWriter.Style().Color.Separator = text.Colors{text.FgBlue}
	}

	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.SetOutputMirror(os.Stdout)

	headerRow := table.Row{"Property", "Value"}
	fields := pt.ShortFields
	if long {
		fields = append(fields, pt.LongFields...)
	}
	tableWriter.AppendHeader(headerRow)
	reflectValue := reflect.ValueOf(pt.Item)
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
			fieldValue = field.Slot(pt.Item)
		} else {
			fieldValue = reflectValue.FieldByName(field.Name)
		}
		tableWriter.AppendRow(table.Row{fieldLabel, fieldValue})
	}
	if pt.Title != "" {
		tableWriter.SetTitle(pt.Title)
		tableWriter.Style().Title.Align = text.AlignCenter
	}
	tableWriter.Render()
}
func (dt PrettyItemTable) PrintJson() {
	output, err := GetIndentJson(dt.Item)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
}

func (dt PrettyItemTable) PrintYaml() {
	output, err := GetYaml(dt.Item)
	if err != nil {
		logging.Fatal("print yaml failed, %s", err)
	}
	fmt.Println(output)
}

func PrintPrettyTableFormat(table PrettyTable, long bool, format string) {
	switch format {
	case TABLE, "default", "":
		table.Print(long)
	case TABLE_LIGHT:
		table.Style = STYLE_LIGHT
		table.Print(long)
	case JSON:
		table.PrintJson()
	case YAML:
		table.PrintYaml()
	default:
		logging.Fatal("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
	}
}
