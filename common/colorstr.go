package common

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

func containsText(stringList []string, s string) bool {
	for _, str := range stringList {
		if s == str {
			return true
		}
	}
	return false
}

func (cf ColorFormater) Format(text string) string {
	switch {
	case containsText(cf.Green, text):
		return GetGreenText(text)
	case containsText(cf.Yellow, text):
		return GetYellowText(text)
	case containsText(cf.Red, text):
		return GetRedText(text)
	default:
		return text
	}
}

var BaseColorFormatter ColorFormater

func init() {
	BaseColorFormatter = ColorFormater{
		Green: []string{
			"enabled", "up", "true", "yes", "success", "active", "ACTIVE",
			"Running", "available", ":)", "Success", "completed",
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
