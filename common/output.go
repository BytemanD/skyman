package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/BytemanD/go-console/console"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/BytemanD/skyman/openstack/model/nova"
)

const (
	DEFAULT     = "default"
	TABLE       = "table"
	TABLE_LIGHT = "table-light"
	JSON        = "json"
	YAML        = "yaml"
)

func GetOutputFormats() []string {
	return []string{TABLE, TABLE_LIGHT, JSON, YAML}
}

func PrintPrettyTable(table PrettyTable, long bool) string {
	return PrintPrettyTableFormat(table, long, CONF.Format)
}

func PrintPrettyItemTable(table PrettyItemTable) string {
	switch CONF.Format {
	case TABLE, "":
		return table.Print(true)
	case TABLE_LIGHT:
		table.Style = STYLE_LIGHT
		return table.Print(true)
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

func PrintAggregate(aggregate nova.Aggregate) {
	pt := PrettyItemTable{
		Item: aggregate,
		ShortFields: []Column{
			{Name: "Id"}, {Name: "Name"}, {Name: "AvailabilityZone"},
			{Name: "Hosts", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.Aggregate)
				return strings.Join(p.Hosts, "\n")
			}},
			{Name: "Metadata", Slot: func(item interface{}) interface{} {
				p, _ := (item).(nova.Aggregate)
				return p.MarshalMetadata()
			}},
			{Name: "CreatedAt"}, {Name: "UpdatedAt"},
			{Name: "Deleted"}, {Name: "DeletedAt"},
		},
	}
	PrintPrettyItemTable(pt)
}

type DataRender[T any] interface {
	SetStyle(style table.Style)
	GetJson() (string, error)
	GetYaml() (string, error)
	Print(more ...bool)
}

func PrintDataTable[T any](dt DataRender[T], long bool) {
	switch CONF.Format {
	case TABLE, "", "default":
		dt.Print(long)
	case TABLE_LIGHT:
		dt.SetStyle(table.StyleLight)
		dt.Print(long)
	case JSON:
		if data, err := dt.GetJson(); err == nil {
			fmt.Println(data)
		} else {
			console.Error("get json failed: %s", err)
			os.Exit(1)
		}
	case YAML:
		if data, err := dt.GetYaml(); err == nil {
			fmt.Println(data)
		} else {
			console.Error("get yaml failed: %s", err)
			os.Exit(1)
		}
	default:
		console.Error("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
		os.Exit(1)
	}
}
