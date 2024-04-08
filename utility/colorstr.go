package utility

import (
	"github.com/BytemanD/easygo/pkg/stringutils"
	"github.com/fatih/color"
)

type ColorFormater struct {
	Green  []string
	Yellow []string
	Red    []string
}

func (cf ColorFormater) Format(text string) string {
	switch {
	case stringutils.ContainsString(cf.Green, text):
		return color.GreenString(text)
	case stringutils.ContainsString(cf.Yellow, text):
		return color.YellowString(text)
	case stringutils.ContainsString(cf.Red, text):
		return color.RedString(text)
	default:
		return text
	}
}

var BaseColorFormatter ColorFormater

func ColorString(text string) string {
	return BaseColorFormatter.Format(text)
}

func init() {
	BaseColorFormatter = ColorFormater{
		Green: []string{
			"enabled", "up", "true", "yes", "success", "active", "ACTIVE",
			"Running", "available", ":)", ":-)", "Success", "completed",
		},
		Yellow: []string{
			"SHUTOFF", "ShutDown", "Unknown", "in-use",
		},
		Red: []string{
			"disabled", "down", "DOWN", "false", "no", "failed", "error",
			"ERROR", "unavailable", "XXX",
		},
	}
}
