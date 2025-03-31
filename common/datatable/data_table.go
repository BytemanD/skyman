package datatable

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/utility"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

func splitTitle(s string) string {
	return strings.Join(lo.Words(s), " ")
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

	count int
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

func (t DataTable[T]) getHeaderRow(columns []Column[T]) table.Row {
	headerRow := table.Row{}
	for _, col := range columns {
		title := col.Text
		if title == "" {
			title = strings.Join(lo.Words(col.Name), " ")
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
func (t *DataTable[T]) SetStyle(style table.Style) {
	t.Style = style
}
func (t DataTable[T]) IsStyleLight() bool {
	return t.Style.Name != table.StyleDefault.Name
}
func (t DataTable[T]) Count() int {
	return t.count
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
func (t DataTable[T]) renderItem(columns []Column[T], item T) table.Row {
	reflectValue := reflect.ValueOf(item)
	row := table.Row{}
	matched := true
	for _, column := range columns {
		var value any
		switch {
		case column.RenderFunc != nil:
			value = column.RenderFunc(item)
		case column.SlotColumn != nil:
			value = column.SlotColumn(item, column)
		case column.Text != "":
			value = column.Text
		default:
			value = reflectValue.FieldByName(column.Name)
		}
		// 匹配
		if len(column.Matchs) > 0 {
			if slice.Contain(column.Matchs, fmt.Sprintf("%v", value)) {
				matched = false
				break
			}
		}
		if t.IsStyleLight() {
			value = utility.NewColorStatus(fmt.Sprint(value)).String()
		}
		row = append(row, value)
	}
	if matched {
		return row
	}
	return nil
}
func (t *DataTable[T]) Render(more ...bool) string {
	tableWriter := t.tableWriter()
	t.count = 0
	columns := t.Columns
	if len(more) > 0 && more[0] {
		columns = append(columns, t.MoreColumns...)
	}
	tableWriter.AppendHeader(t.getHeaderRow(columns))
	for _, item := range t.Items {
		row := t.renderItem(columns, item)
		if row != nil {
			tableWriter.AppendRow(row)
			t.count += 1
		}
	}
	return tableWriter.Render()
}

func (t DataTable[T]) Print(more ...bool) {
	fmt.Println(t.Render(more...))
	fmt.Println("Items: ", t.Count())
}
