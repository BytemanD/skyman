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
	return []string{DEFAULT, TABLE, TABLE_LIGHT, JSON, YAML}
}

func PrintPrettyTable(table PrettyTable, long bool) {
	switch CONF.Format {
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
func PrintDataTable(table DataTable) {
	switch CONF.Format {
	case TABLE, "default", "":
		table.Print(true)
	case TABLE_LIGHT:
		table.Style = STYLE_LIGHT
		table.Print(true)
	case JSON:
		table.PrintJson()
	case YAML:
		table.PrintYaml()
	default:
		logging.Fatal("invalid output format: %s, valid formats: %v", CONF.Format,
			GetOutputFormats())
	}
}
