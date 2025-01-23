package datatable

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/BytemanD/skyman/utility"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_ITERATOR_COLUMN_MAX_WIDTH = 100
)

type DataIterator[T any] struct {
	Items               []T
	Fields              []Field[T]
	MoreFields          []Field[T]
	ValueColumnMaxWidth int
	Title               string
	Style               table.Style
	AutoIndex           bool
	SortBy              []table.SortBy
	SeparateRows        bool
	Output              io.Writer
	count               int
}

func (t DataIterator[T]) tableWriter() table.Writer {
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

func (t DataIterator[T]) getHeaderRow() table.Row {
	return table.Row{"Property", "Value"}
}

func (t DataIterator[T]) getColumnConfigs() []table.ColumnConfig {
	valueColumneWdith := utility.OneOfNumber(t.ValueColumnMaxWidth, DEFAULT_ITERATOR_COLUMN_MAX_WIDTH)
	return []table.ColumnConfig{{Number: 2, WidthMax: valueColumneWdith}}
}

func (t *DataIterator[T]) AddItems(items []T) *DataIterator[T] {
	t.Items = append(t.Items, items...)
	return t
}

func (t *DataIterator[T]) SetStyle(style table.Style) {
	t.Style = style
}

func (t DataIterator[T]) IsStyleLight() bool {
	return t.Style.Name != table.StyleDefault.Name
}
func (t DataIterator[T]) Count() int {
	return t.count
}
func (t DataIterator[T]) GetJson() (string, error) {
	return stringutils.JsonDumpsIndent(t.Items)
}
func (t DataIterator[T]) GetYaml() (string, error) {
	bytes, err := yaml.Marshal(t.Items)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
func (t DataIterator[T]) renderItem(item T, fields []Field[T]) string {
	tableWriter := t.tableWriter()
	tableWriter.AppendHeader(t.getHeaderRow())

	reflectValue := reflect.ValueOf(item)
	matched := true
	for _, field := range fields {
		var fieldText string
		if field.Text != "" {
			fieldText = field.Text
		} else {
			fieldText = splitTitle(field.Name)
		}

		var fieldValue interface{}
		switch {
		case field.RenderFunc != nil:
			fieldValue = field.RenderFunc(item)
		default:
			reflectField := reflectValue.FieldByName(field.Name)
			if field.Marshal {
				j, _ := json.Marshal(reflectField.Interface())
				fieldValue = string(j)
			} else {
				fieldValue = reflectField
			}
		}
		// 匹配
		if len(field.Matchs) > 0 {
			if slice.Contain(field.Matchs, fmt.Sprintf("%v", fieldValue)) {
				matched = false
				break
			}
		}

		if t.IsStyleLight() {
			fieldValue = utility.NewColorStatus(fmt.Sprint(fieldValue)).String()
		}
		tableWriter.AppendRow(table.Row{fieldText, fieldValue})
	}
	if matched {
		return tableWriter.Render()
	}
	return ""
}
func (t *DataIterator[T]) Render(more ...bool) []string {
	t.count = 0
	fields := t.Fields
	if len(more) > 0 && more[0] {
		fields = append(fields, t.MoreFields...)
	}
	data := []string{}
	for _, item := range t.Items {
		itemData := t.renderItem(item, fields)
		if itemData != "" {
			data = append(data, itemData)
			t.count += 1
		}
	}
	return data
}

func (t DataIterator[T]) Print(more ...bool) {
	data := t.Render(more...)
	for _, table := range data {
		fmt.Println(table)
	}
	if t.Count() > 1 {
		fmt.Println("Items: ", t.Count())
	}
}
