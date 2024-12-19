package internal

import (
	"net/url"

	"github.com/go-resty/resty/v2"
)

type ResourceApi struct {
	Client  *resty.Client
	BaseUrl string

	EnableAllTenant bool
	SingularKey     string
	PluralKey       string
}

func (r ResourceApi) NewRequest(method string, u string, q url.Values, body interface{}, result interface{}) *resty.Request {
	req := r.Client.R().SetQueryParamsFromValues(q).SetResult(result)
	if reqUrl, err := url.JoinPath(r.BaseUrl, u); err != nil {
		req.Error = err
	} else {
		req.URL = reqUrl
	}
	req.Method = method
	if body != nil {
		req.SetBody(body)
	}
	if url, err := url.JoinPath(r.BaseUrl, u); err != nil {
		req.Error = err
	} else {
		req.URL = url
	}
	return req
}
func (r ResourceApi) NewGetRequest(u string, q url.Values, result interface{}) *resty.Request {
	return r.NewRequest(resty.MethodGet, u, q, nil, result)
}
func (r ResourceApi) NewDeleteRequest(u string, q url.Values, result interface{}) *resty.Request {
	return r.NewRequest(resty.MethodDelete, u, q, nil, result)
}
func (r ResourceApi) Get(url string, q url.Values, result interface{}) (*resty.Response, error) {
	return r.NewGetRequest(url, q, result).Send()
}
func (r ResourceApi) Delete(url string) (*resty.Response, error) {
	return r.NewDeleteRequest(url, nil, nil).Send()
}
