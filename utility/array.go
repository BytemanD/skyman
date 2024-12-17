package utility

import (
	"net/url"
)

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

type StringRing struct {
	Items []string
	index int
}

func (r *StringRing) Next() string {
	if r.Items == nil || len(r.Items) == 0 {
		return ""
	}
	if r.index >= len(r.Items) {
		r.index = 0
	}
	s := r.Items[r.index]
	return s
}
func (r *StringRing) Sample(count int) []string {
	sample := []string{}
	for i := 0; i < count; i++ {
		sample = append(sample, r.Next())
	}
	return sample
}

func Filter[T any](items []T, filterFun func(x T) bool) []T {
	filterd := []T{}
	for _, item := range items {
		if filterFun(item) {
			filterd = append(filterd, item)
		}
	}
	return filterd
}
func OneOfString(strs ...string) string {
	if len(strs) == 0 {
		return ""
	}
	for _, str := range strs {
		if str != "" {
			return str
		}
	}
	return strs[len(strs)-1]
}

func OneOfNumber[T int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | float32 | float64](values ...T) T {
	if len(values) == 0 {
		return 0
	}
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return values[len(values)-1]
}
func OneOfBoolean(values ...bool) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if value {
			return value
		}
	}
	return values[len(values)-1]
}
func OneOfStringArrays(arraysList ...[]string) []string {
	if len(arraysList) == 0 {
		return []string{}
	}
	for _, arrays := range arraysList {
		if len(arrays) != 0 {
			return arrays
		}
	}
	return arraysList[len(arraysList)-1]
}
