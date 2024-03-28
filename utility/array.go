package utility

import "net/url"

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
