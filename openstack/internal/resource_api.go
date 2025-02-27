package internal

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
)

func checkError(resp *resty.Response, err error) (*session.Response, error) {
	if err != nil || resp == nil {
		return &session.Response{Response: resp}, err
	}
	if resp.IsError() {
		return &session.Response{Response: resp}, session.HttpError{
			Status:  resp.StatusCode(),
			Reason:  resp.Status(),
			Message: string(resp.Body()),
		}
	}
	return &session.Response{Response: resp}, nil
}

type ResourceApi struct {
	Client      *resty.Client
	BaseUrl     string
	ResourceUrl string

	EnableAllTenant bool
	SingularKey     string
	PluralKey       string

	MicroVersion *model.ApiVersion
}

func (c *ResourceApi) MicroVersionLargeEqual(version string) bool {
	clientVersion := getMicroVersion(c.MicroVersion.Version)
	otherVersion := getMicroVersion(version)
	if clientVersion.Version > otherVersion.Version {
		return true
	} else if clientVersion.Version == otherVersion.Version {
		return clientVersion.MicroVersion >= otherVersion.MicroVersion
	} else {
		return false
	}
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
func (r *ResourceApi) R() *session.Request {
	return &session.Request{
		Request:     r.Client.R(),
		Baseurl:     r.BaseUrl,
		ResourceUrl: r.ResourceUrl,
	}
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

func (r ResourceApi) Get(url string, q url.Values, result interface{}) (*session.Response, error) {
	resp, err := r.NewGetRequest(url, q, result).Send()
	return checkError(resp, err)
}
func (r ResourceApi) Delete(u string, query ...url.Values) (*session.Response, error) {
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
func (r ResourceApi) ResourceDelete(id string, query ...url.Values) (*session.Response, error) {
	if r.ResourceUrl == "" {
		return nil, fmt.Errorf("ResourceUrl is empty")
	}
	return r.Delete(r.ResourceUrl+"/"+id, query...)
}
func (r ResourceApi) Post(url string, body interface{}, result interface{}) (*session.Response, error) {
	return checkError(
		r.NewPostRequest(url, body, result).Send(),
	)
}
func (r ResourceApi) Put(url string, body interface{}, result interface{}) (*session.Response, error) {
	return checkError(
		r.NewPutRequest(url, body, result).Send(),
	)
}
func (r ResourceApi) Patch(url string, body interface{}, result interface{}, headers map[string]string) (*session.Response, error) {
	return checkError(
		r.NewPatchRequest(url, body, result).SetHeaders(headers).Send(),
	)
}
func ListResource[T any](r ResourceApi, query url.Values, detail ...bool) ([]T, error) {
	if r.ResourceUrl == "" {
		return nil, fmt.Errorf("ResourceUrl is empty")
	}
	if r.PluralKey == "" {
		return nil, fmt.Errorf("PluralKey is empty")
	}
	items := []T{}
	respBody := map[string]interface{}{}
	if len(detail) > 0 && detail[0] {
		if _, err := r.R().SetQuery(query).SetResult(&respBody).Get("detail"); err != nil {
			return nil, err
		}
	} else {
		if _, err := r.R().SetQuery(query).SetResult(&respBody).Get(); err != nil {
			return nil, err
		}
	}
	itemsData, err := json.Marshal(respBody[r.PluralKey])
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(itemsData, &items); err != nil {
		return nil, err
	}
	return items, nil
}
func ShowResource[T any](r ResourceApi, id string) (*T, error) {
	if r.ResourceUrl == "" {
		return nil, fmt.Errorf("ResourceUrl is empty")
	}
	if r.SingularKey == "" {
		return nil, fmt.Errorf("SingularKey is empty")
	}
	respBody := map[string]*T{}
	// respBody[r.SingularKey] = T{}
	if _, err := r.R().SetResult(&respBody).Get(id); err != nil {
		return nil, err
	}
	return respBody[r.SingularKey], nil
}

func DeleteResource(r ResourceApi, id string, query ...url.Values) (*session.Response, error) {
	if r.ResourceUrl == "" {
		return nil, fmt.Errorf("ResourceUrl is empty")
	}
	return r.Delete(r.ResourceUrl+"/"+id, query...)
}

func FindResource[T any](
	idOrName string,
	showFunc func(id string) (*T, error),
	listFunc func(query url.Values) ([]T, error),
	allTenants ...bool,
) (*T, error) {
	t, err := showFunc(idOrName)
	if err == nil {
		return t, nil
	}
	switch errType := err.(type) {
	case session.HttpError:
		if !errType.IsNotFound() {
			return nil, err
		}
	default:
		return nil, err
	}
	query := url.Values{"name": []string{idOrName}}
	if len(allTenants) > 0 && allTenants[0] {
		query.Set("all_tenants", "1")
	}
	ts, err := listFunc(query)
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
