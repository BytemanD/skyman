package utility

import "net/url"

type ObjectArray []interface{}

func CopyToInterfaceSlice(list ObjectArray) []interface{} {
	items := []interface{}{}
	// for _, i := range list {
	// 	items = append(items, i)
	// }
	items = append(items, list...)
	return items
}

// [min], max, [step]
func Range(args ...int) []int {
	var min, max, step int
	switch len(args) {
	case 0:
		min, max, step = 0, 0, 1
	case 1:
		min, max, step = 0, args[0], 1
	case 2:
		min, max, step = args[0], args[1], 1
	default:
		min, max, step = args[0], args[1], args[2]
	}

	items := []int{}
	for i := min; i < max; i += step {
		items = append(items, i)
	}
	return items
}

func UrlValues(m map[string]string) url.Values {
	query := url.Values{}
	for k, v := range m {
		if v == "" {
			continue
		}
		query.Set(k, v)
	}
	return query
}
