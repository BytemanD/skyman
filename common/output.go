package common

import "github.com/BytemanD/easygo/pkg/global/logging"

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
		logging.Fatal("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
	}
	return ""
}
