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
	RenderFunc func(item T) any
	SlotColumn func(item T, column Column[T]) any
	SortMode   table.SortMode
	Matchs     []string
	Marshal    bool
	WidthMax   int
	Align      text.Align
}

type Field[T any] struct {
	Name string
	Text string
	// 只有 Table.Style 等于 light 是才会生效
	AutoColor  bool
	ForceColor bool
	RenderFunc func(item T) any
	SlotColumn func(item T, column Field[T]) any
	SortMode   table.SortMode
	// 模糊匹配
	Matchs   []string
	Marshal  bool
	WidthMax int
	Align    text.Align
}
