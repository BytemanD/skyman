package utility

type ObjectArray []interface{}

func CopyToInterfaceSlice(list ObjectArray) []interface{} {
	items := []interface{}{}
	// for _, i := range list {
	// 	items = append(items, i)
	// }
	items = append(items, list...)
	return items
}
