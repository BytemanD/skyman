package cli

import (
	"fmt"
)

// TODO: use easygo

func GetRedText(text string) string {
	return fmt.Sprintf("\033[1;31;40m%s\033[0m", text)
}
func GetGreenText(text string) string {
	return fmt.Sprintf("\033[1;32;40m%s\033[0m", text)
}
func GetYellowText(text string) string {
	return fmt.Sprintf("\033[1;33;40m%s\033[0m", text)
}

type ColorFormater struct {
	Green  []string
	Yellow []string
	Red    []string
}

func (cf ColorFormater) Format(text string) string {
	for _, str := range cf.Green {
		if text == str {
			return GetGreenText(text)
		}
	}
	for _, str := range cf.Yellow {
		if text == str {
			return GetYellowText(text)
		}
	}
	for _, str := range cf.Red {
		if text == str {
			return GetRedText(text)
		}
	}
	return text
}

var BaseColorFormatter ColorFormater

func init() {
	BaseColorFormatter = ColorFormater{
		Green: []string{
			"enabled", "up", "true", "yes", "success", "active", "ACTIVE",
			"Running", "available", ":)",
		},
		Yellow: []string{
			"SHUTOFF", "ShutDown", "Unknown",
		},
		Red: []string{
			"disabled", "down", "false", "no", "failed", "error",
			"ERROR", "unavailable", "XXX",
		},
	}
}
