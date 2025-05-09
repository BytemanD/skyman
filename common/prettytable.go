package common

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/samber/lo"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	STYLE_LIGHT = "light"
)

type Column struct {
	Name string
	Text string
	// 只有 Table.Style 等于 light 是才会生效
	AutoColor  bool
	ForceColor bool
	Slot       func(item any) any
	SlotColumn func(item any, column Column) any
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
	Items             []any
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
		capColName := lo.Capitalize(colName)
		if slice.Contain(pt.DisplayFields, capColName) {
			continue
		}
		pt.DisplayFields = append(pt.DisplayFields, capColName)
	}
}
func (pt *PrettyTable) AddItems(items any) {
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
		pt.Items = []any{}
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
	}
	headerRow := table.Row{}

	columns := []Column{}
	if len(pt.DisplayFields) > 0 {
		for _, field := range pt.DisplayFields {
			found := lo.Filter(
				append(pt.ShortColumns, pt.LongColumns...),
				func(x Column, index int) bool { return x.Name == field },
			)
			if len(found) > 0 {
				columns = append(columns, found...)
			} else {
				columns = append(columns, Column{Name: field})
			}
		}
	} else {
		columns = pt.ShortColumns
		if len(pt.DisplayFields) == 0 {
			if long {
				columns = append(columns, pt.LongColumns...)
			}
		}
	}

	sortBy := []table.SortBy{}
	for _, column := range columns {
		var title string
		if column.Text == "" {
			title = strings.Join(lo.Words(column.Name), " ")
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
			var value any
			if column.Slot != nil {
				value = column.Slot(item)
			} else if column.SlotColumn != nil {
				value = column.SlotColumn(item, column)
			} else {
				value = reflectValue.FieldByName(column.Name)
			}
			// match filter
			if len(column.Filters) > 0 {
				if !slice.Contain(column.Filters, fmt.Sprintf("%v", value)) {
					isFiltered = true
					break
				}
			}
			if pt.Search != "" && !strings.Contains(fmt.Sprintf("%v", value), pt.Search) {
				matchedCount -= 1
			}
			if column.ForceColor || (column.AutoColor && pt.Style == STYLE_LIGHT) {
				value = utility.NewColorStatus(fmt.Sprint(value))
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
		console.Fatal("print json failed, %s", err)

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
		console.Fatal("print json failed, %s", err)

	}
	fmt.Println(output)
	return output
}

type PrettyItemTable struct {
	ShortFields     []Column
	LongFields      []Column
	Item            any
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
			fieldValue any
			fieldLabel string
		)
		if field.Text == "" {
			fieldLabel = strings.Join(lo.Words(field.Name), " ")
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
				if field.AutoColor {
					fieldValue = utility.NewColorStatus(fmt.Sprint(fieldValue))
				}
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
		console.Fatal("print json failed, %s", err)

	}
	fmt.Println(output)
	return output
}

func (dt PrettyItemTable) PrintYaml() string {
	output, err := GetYaml(dt.Item)
	if err != nil {
		console.Fatal("print yaml failed, %s", err)

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
		console.Error("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
		os.Exit(1)
	}
	return ""
}
