package datatable

import (
	"fmt"
	"io"
	"reflect"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/utility"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
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

type DataTable[T any] struct {
	Items       []T
	Columns     []Column[T]
	MoreColumns []Column[T]
	Title       string

	Style        table.Style
	AutoIndex    bool
	SortBy       []table.SortBy
	SeparateRows bool
	Output       io.Writer
}

func (t DataTable[T]) tableWriter() table.Writer {
	tableWriter := table.NewWriter()
	if t.IsStyleLight() {
		tableWriter.SetStyle(table.StyleLight)
		tableWriter.Style().Color.Header = text.Colors{text.FgBlue, text.Bold}
		tableWriter.Style().Color.Border = text.Colors{text.FgBlue}
		tableWriter.Style().Color.Separator = text.Colors{text.FgBlue}
	}
	tableWriter.Style().Format.Header = text.FormatDefault
	tableWriter.Style().Options.SeparateRows = t.SeparateRows

	if t.Title != "" {
		tableWriter.SetTitle(t.Title)
	}
	tableWriter.SortBy(t.SortBy)
	tableWriter.SetAutoIndex(t.AutoIndex)
	tableWriter.SetColumnConfigs(t.getColumnConfigs())

	return tableWriter
}

func (t DataTable[T]) getHeaderRow(more bool) table.Row {
	headerRow := table.Row{}
	columns := t.Columns
	if more {
		columns = append(columns, t.MoreColumns...)
	}
	for _, col := range columns {
		var title string
		if col.Text != "" {
			title = col.Text
		} else {
			title = splitTitle(col.Name)
		}
		headerRow = append(headerRow, title)
	}
	return headerRow
}

func (t DataTable[T]) getColumnConfigs() []table.ColumnConfig {
	colConfigs := []table.ColumnConfig{}
	for i, column := range t.Columns {
		colConfigs = append(colConfigs, table.ColumnConfig{
			Number:   i + 1,
			WidthMax: column.WidthMax,
			Align:    column.Align,
		})
	}
	return colConfigs
}

func (t *DataTable[T]) AddItems(items []T) *DataTable[T] {
	t.Items = append(t.Items, items...)
	return t
}
func (t *DataTable[T]) AddColumns(columns []Column[T]) *DataTable[T] {
	t.Columns = append(t.Columns, columns...)
	return t
}
func (t *DataTable[T]) SetStyle(style table.Style) *DataTable[T] {
	t.Style = style
	return t
}
func (t DataTable[T]) IsStyleLight() bool {
	return t.Style.Name != table.StyleDefault.Name
}
func (t DataTable[T]) ItemsNum() int {
	return len(t.Items)
}
func (t DataTable[T]) GetJson() (string, error) {
	return stringutils.JsonDumpsIndent(t.Items)
}
func (t DataTable[T]) GetYaml() (string, error) {
	bytes, err := yaml.Marshal(t.Items)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func (t DataTable[T]) Render(more ...bool) string {
	tableWriter := t.tableWriter()

	columns := t.Columns
	if len(more) > 0 && more[0] {
		tableWriter.AppendHeader(t.getHeaderRow(true))
		columns = append(columns, t.MoreColumns...)
	} else {
		tableWriter.AppendHeader(t.getHeaderRow(false))
	}
	fmt.Println(more, len(columns))
	for _, item := range t.Items {
		reflectValue := reflect.ValueOf(item)
		row := table.Row{}

		isFiltered := false
		matchedCount := len(columns)
		for _, column := range columns {
			var value interface{}
			switch {
			case column.RenderFunc != nil:
				value = column.RenderFunc(item)
			case column.SlotColumn != nil:
				value = column.SlotColumn(item, column)
			default:
				value = reflectValue.FieldByName(column.Name)
			}
			// match filter
			if len(column.Filters) > 0 {
				if !stringutils.ContainsString(column.Filters, fmt.Sprintf("%v", value)) {
					isFiltered = true
					break
				}
			}
			// if t.Search != "" && !strings.Contains(fmt.Sprintf("%v", value), pt.Search) {
			// 	matchedCount -= 1
			// }
			if t.IsStyleLight() {
				value = utility.NewColorStatus(fmt.Sprint(value)).String()
			}
			row = append(row, value)
		}
		if isFiltered || matchedCount <= 0 {
			continue
		}
		tableWriter.AppendRow(row)
	}
	return tableWriter.Render()
}

func (t DataTable[T]) Print(more ...bool) {
	fmt.Println(t.Render(more...))
	fmt.Println("Items: ", len(t.Items))
}
