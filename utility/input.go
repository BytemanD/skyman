package utility

import (
	"fmt"
	"strings"
)

func ScanfComfirm(message string, yes, no []string) bool {
	var confirm string
	for {
		fmt.Printf("%s [%s|%s]: ", message, strings.Join(yes, "/"), strings.Join(no, "/"))
		fmt.Scanf("%s %d %f", &confirm)
		if StringsContain(yes, confirm) {
			return true
		} else if StringsContain(no, confirm) {
			return false
		} else {
			fmt.Print("输入错误, 请重新输入!")
		}
	}
}
