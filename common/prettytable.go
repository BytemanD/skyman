package common

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/utility"
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
	SlotColumn func(item interface{}, column Column) interface{}
	Sort       bool
	SortMode   table.SortMode
	Filters    []string
	Marshal    bool
	WidthMax   int
	Align      text.Align
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
	Filters           map[string]string
	Search            string
	DisplayFields     []string
}

func (pt *PrettyTable) AddDisplayFields(fields ...string) {
	if pt.DisplayFields == nil {
		pt.DisplayFields = []string{}
	}
	pt.DisplayFields = []string{"Id"}
	for _, colName := range fields {
		if stringutils.ContainsString(pt.DisplayFields, colName) {
			continue
		}
		pt.DisplayFields = append(pt.DisplayFields, colName)
	}
}
func (pt *PrettyTable) AddItems(items interface{}) {
	value := reflect.ValueOf(items)
	for i := 0; i < value.Len(); i++ {
		if value.Index(i).Kind() == reflect.Ptr {
			pt.Items = append(pt.Items, value.Index(i).Elem().Interface())
		} else {
			pt.Items = append(pt.Items, value.Index(i).Interface())
		}
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
func (pt PrettyTable) GetShortColumnIndex(column string) int {
	for i, c := range pt.ShortColumns {
		if c.Name == column {
			return i
		}
	}
	return -1
}
func (pt PrettyTable) GetLongColumnIndex(column string) int {
	for i, c := range pt.LongColumns {
		if c.Name == column {
			return i
		}
	}
	return -1
}
func (pt PrettyTable) RenderToTable(long bool) string {
	tableWriter := pt.getTableWriter()
	if pt.Title != "" {
		fmt.Println(pt.Title)
		// tableWriter.SetTitle("%s", pt.Title)
		// tableWriter.Style().Title.Align = text.AlignCenter
	}
	headerRow := table.Row{}

	columns := []Column{}
	if pt.DisplayFields != nil && len(pt.DisplayFields) > 0 {
		for _, field := range pt.DisplayFields {
			found := utility.Filter(
				append(pt.ShortColumns, pt.LongColumns...),
				func(x Column) bool {
					return x.Name == field
				},
			)
			if len(found) > 0 {
				columns = append(columns, found...)
			} else {
				columns = append(columns, Column{Name: field})
			}
		}
	} else {
		columns = pt.ShortColumns
		if pt.DisplayFields == nil || len(pt.DisplayFields) == 0 {
			if long {
				columns = append(columns, pt.LongColumns...)
			}
		}
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
	colConfigs := []table.ColumnConfig{}
	for i, column := range columns {
		colConfigs = append(colConfigs, table.ColumnConfig{
			Number:   i + 1,
			WidthMax: column.WidthMax,
			Align:    column.Align,
		})
	}
	tableWriter.SetColumnConfigs(colConfigs)

	for _, item := range pt.Items {
		reflectValue := reflect.ValueOf(item)
		row := table.Row{}
		isFiltered := false
		matchedCount := len(columns)
		for _, column := range columns {
			var value interface{}
			if column.Slot != nil {
				value = column.Slot(item)
			} else if column.SlotColumn != nil {
				value = column.SlotColumn(item, column)
			} else {
				value = reflectValue.FieldByName(column.Name)
			}
			// match filter
			if len(column.Filters) > 0 {
				if !stringutils.ContainsString(column.Filters, fmt.Sprintf("%v", value)) {
					isFiltered = true
					break
				}
			}
			if pt.Search != "" && !strings.Contains(fmt.Sprintf("%v", value), pt.Search) {
				matchedCount -= 1
			}
			if column.ForceColor || (column.AutoColor && pt.Style == STYLE_LIGHT) {
				value = utility.ColorString(fmt.Sprint(value))
			}
			row = append(row, value)
		}
		if isFiltered || matchedCount <= 0 {
			continue
		}
		tableWriter.AppendRow(row)
	}

	// TODO: 当前只能按Columns 顺序排序
	tableWriter.SortBy(sortBy)
	return tableWriter.Render()
}
func (pt PrettyTable) Print(long bool) string {
	result := pt.RenderToTable(long)
	// tableWriter.Render()
	if !pt.HideTotalItems {
		fmt.Printf("Total items: %d\n", len(pt.Items))
	}
	return result
}
func (pt PrettyTable) RenderToJson() (string, error) {
	return stringutils.JsonDumpsIndent(pt.Items)
}
func (pt PrettyTable) PrintJson() string {
	output, err := pt.RenderToJson()
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
	return output
}
func (pt PrettyTable) RenderToYaml() (string, error) {
	return GetYaml(pt.Items)
}
func (pt PrettyTable) PrintYaml() string {
	output, err := pt.RenderToYaml()
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
	return output
}

type PrettyItemTable struct {
	ShortFields     []Column
	LongFields      []Column
	Item            interface{}
	Title           string
	Style           string
	Number2WidthMax int
}

func (pt PrettyItemTable) Print(long bool) string {
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
	if pt.Number2WidthMax == 0 {
		tableWriter.SetColumnConfigs([]table.ColumnConfig{
			{Number: 2, WidthMax: 100},
		})
	} else {
		tableWriter.SetColumnConfigs([]table.ColumnConfig{
			{Number: 2, WidthMax: pt.Number2WidthMax},
		})
	}
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
			reflectField := reflectValue.FieldByName(field.Name)
			if field.Marshal {
				j, _ := json.Marshal(reflectField.Interface())
				fieldValue = string(j)
			} else {
				fieldValue = reflectField
			}
		}
		tableWriter.AppendRow(table.Row{fieldLabel, fieldValue})
	}
	if pt.Title != "" {
		tableWriter.SetTitle(pt.Title)
		tableWriter.Style().Title.Align = text.AlignCenter
	}
	return tableWriter.Render()
}
func (dt PrettyItemTable) PrintJson() string {
	output, err := stringutils.JsonDumpsIndent(dt.Item)
	if err != nil {
		logging.Fatal("print json failed, %s", err)
	}
	fmt.Println(output)
	return output
}

func (dt PrettyItemTable) PrintYaml() string {
	output, err := GetYaml(dt.Item)
	if err != nil {
		logging.Fatal("print yaml failed, %s", err)
	}
	fmt.Println(output)
	return output
}

func PrintPrettyTableFormat(table PrettyTable, long bool, format string) string {
	switch format {
	case TABLE, "default", "":
		table.Print(long)
	case TABLE_LIGHT:
		table.Style = STYLE_LIGHT
		return table.Print(long)
	case JSON:
		return table.PrintJson()
	case YAML:
		return table.PrintYaml()
	default:
		logging.Fatal("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
	}
	return ""
}
