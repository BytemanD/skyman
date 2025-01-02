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

type ColorStatus struct {
	status   string
	Formater ColorFormater
}

func (s ColorStatus) String() string {
	switch {
	case stringutils.ContainsString(s.Formater.Green, s.status):
		return color.GreenString(s.status)
	case stringutils.ContainsString(s.Formater.Yellow, s.status):
		return color.YellowString(s.status)
	case stringutils.ContainsString(s.Formater.Red, s.status):
		return color.RedString(s.status)
	default:
		return s.status
	}
}

func NewColorStatus(s string) ColorStatus {
	return ColorStatus{
		Formater: ColorFormater{
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
		},
		status: s,
	}
}
