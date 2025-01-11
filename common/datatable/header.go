package datatable

import (
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

type Column[T any] struct {
	Name string
	Text string
	// 只有 Table.Style 等于 light 是才会生效
	AutoColor  bool
	ForceColor bool
	RenderFunc func(item T) interface{}
	SlotColumn func(item T, column Column[T]) interface{}
	SortMode   table.SortMode
	Filters    []string
	Marshal    bool
	WidthMax   int
	Align      text.Align
}
