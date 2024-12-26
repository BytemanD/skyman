package internal

import (
	"fmt"
	"net/url"
	"reflect"

	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

func checkError(resp *resty.Response, err error) (*resty.Response, error) {
	if err != nil || resp == nil {
		return resp, err
	}
	if resp.IsError() {
		return resp, httpclient.HttpError{
			Status:  resp.StatusCode(),
			Reason:  resp.Status(),
			Message: string(resp.Body()),
		}
	}
	return resp, nil
}

type ResourceApi struct {
	Client      *resty.Client
	BaseUrl     string
	ResourceUrl string

	EnableAllTenant bool
	SingularKey     string
	PluralKey       string

	URL_LIST string
	URL_SHOW string
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
func (r ResourceApi) NewPostRequest(u string, body interface{}, result interface{}) *resty.Request {
	return r.NewRequest(resty.MethodPost, u, nil, body, result)
}
func (r ResourceApi) NewPutRequest(u string, body interface{}, result interface{}) *resty.Request {
	return r.NewRequest(resty.MethodPut, u, nil, body, result)
}
func (r ResourceApi) NewPatchRequest(u string, body interface{}, result interface{}) *resty.Request {
	return r.NewRequest(resty.MethodPatch, u, nil, body, result)
}

func (r ResourceApi) Get(url string, q url.Values, result interface{}) (*resty.Response, error) {
	resp, err := r.NewGetRequest(url, q, result).Send()
	return checkError(resp, err)
}
func (r ResourceApi) Delete(u string, query ...url.Values) (*resty.Response, error) {
	var q url.Values
	if len(query) > 0 {
		q = query[0]
	} else {
		q = nil
	}
	return checkError(
		r.NewDeleteRequest(u, q, nil).Send(),
	)
}
func (r ResourceApi) ResourceDelete(id string, query ...url.Values) (*resty.Response, error) {
	if r.ResourceUrl == "" {
		return nil, fmt.Errorf("ResourceUrl is empty")
	}
	return r.Delete(r.ResourceUrl+"/"+id, query...)
}
func (r ResourceApi) Post(url string, body interface{}, result interface{}) (*resty.Response, error) {
	return checkError(
		r.NewPostRequest(url, body, result).Send(),
	)
}
func (r ResourceApi) Put(url string, body interface{}, result interface{}) (*resty.Response, error) {
	return checkError(
		r.NewPutRequest(url, body, result).Send(),
	)
}
func (r ResourceApi) Patch(url string, body interface{}, result interface{}, headers map[string]string) (*resty.Response, error) {
	return checkError(
		r.NewPatchRequest(url, body, result).SetHeaders(headers).Send(),
	)
}
func FindResource[T any](
	idOrName string,
	showFunc func(id string) (*T, error),
	listFunc func(query url.Values) ([]T, error),
) (*T, error) {
	t, err := showFunc(idOrName)
	if err == nil {
		return t, nil
	}
	if _, ok := err.(httpclient.HttpError); !ok {
		return nil, err
	}
	if httpError, _ := err.(httpclient.HttpError); !httpError.IsNotFound() {
		return nil, err
	}
	ts, err := listFunc(url.Values{"name": []string{idOrName}})
	if err != nil {
		return nil, err
	}
	switch len(ts) {
	case 0:
		return nil, fmt.Errorf("resource %s not found", idOrName)
	case 1:
		t := ts[0]
		value := reflect.ValueOf(t)
		valueName := value.FieldByName("Name")
		if valueName.String() != idOrName {
			return nil, fmt.Errorf("resource %s not found", idOrName)
		} else {
			return &t, nil
		}
	default:
		fileted := []T{}
		for _, t := range ts {
			value := reflect.ValueOf(t)
			valueName := value.FieldByName("Name")
			if valueName.Kind() == reflect.String && valueName.String() == idOrName {
				fileted = append(fileted, t)
			}
		}
		if len(fileted) == 0 {
			return nil, fmt.Errorf("resource %s not found", idOrName)
		}
		if len(fileted) == 1 {
			return &fileted[0], nil
		}
		return nil, fmt.Errorf("found %d resources with name %s ", len(fileted), idOrName)
	}
}
