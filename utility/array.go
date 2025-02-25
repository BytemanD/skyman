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
	if len(r.Items) == 0 {
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
