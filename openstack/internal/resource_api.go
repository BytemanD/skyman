package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/session"
	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

func checkError(resp *resty.Response, err error) (*session.Response, error) {
	if err != nil || resp == nil {
		return &session.Response{Response: resp}, err
	}
	if resp.IsError() {
		switch resp.StatusCode() {
		case 404:
			return nil, fmt.Errorf("%w: %w, %s", session.ErrHTTPStatus, session.ErrHTTP404, resp.Body())
		default:
			return nil, fmt.Errorf("%w: [%d], %s", session.ErrHTTPStatus, resp.StatusCode(), string(resp.Body()))
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
	req := r.Client.R().SetQueryParamsFromValues(q).SetResult(result).SetBody(body)
	req.Method = method
	req.URL = u
	fmt.Println("xxxx", r.Client.BaseURL, req.URL)
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
	return r.R().SetResult(&result).SetQuery(q).Get(url)
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

type FindCaseInterface[T any] interface {
	List(query url.Values) ([]T, error)
	Show(id string) (*T, error)
}

func FindIdOrName[T any](api FindCaseInterface[T], idOrName string, allTenants ...bool) (*T, error) {
	if found, err := api.Show(idOrName); err == nil {
		return found, nil
	} else if !errors.Is(err, session.ErrHTTP404) && !errors.Is(err, ErrResourceNotFound) {
		return nil, err
	}
	query := url.Values{"name": []string{idOrName}}
	if len(allTenants) > 0 && allTenants[0] {
		query.Set("all_tenants", "1")
	}
	if founds, err := api.List(query); err != nil {
		return nil, err
	} else {
		fileted := lo.Filter(founds, func(item T, _ int) bool {
			valueName := reflect.ValueOf(item).FieldByName("Name")
			return valueName.Kind() == reflect.String && valueName.String() == idOrName
		})
		switch len(fileted) {
		case 0:
			return nil, fmt.Errorf("%w with id or name %s", ErrResourceNotFound, idOrName)
		case 1:
			return &fileted[0], nil
		default:
			return nil, fmt.Errorf("%w with id or name: %s", ErrResourceMulti, idOrName)
		}
	}
}
