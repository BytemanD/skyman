package result

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/BytemanD/skyman/openstack/session"
)

type HttpResult struct {
	Resp *session.Response
	Err  error
}

func (result HttpResult) RequestId() string {
	if result.Resp == nil {
		return ""
	}
	if values, ok := result.Resp.Header()[session.HEADER_REQUEST_ID]; !ok {
		return ""
	} else {
		return values[0]
	}
}
func (result HttpResult) IsError() bool {
	if result.Err != nil {
		return true
	}
	return result.Resp.IsError()
}
func (result HttpResult) GetError() error {
	if !result.IsError() {
		return nil
	}
	return fmt.Errorf("result Err=%v, StatusCode: %v",
		result.Err, result.Resp.StatusCode())
}
func (result HttpResult) NotFound() bool {
	if !result.IsError() {
		return false
	}
	return result.Resp.StatusCode() == http.StatusNotFound
}
func (result HttpResult) StringBody() string {
	return string(result.Resp.Body())
}

type ItemsResult[T any] struct {
	HttpResult
	Key string
}

func (result *ItemsResult[T]) SetKey(key string) ItemsResult[T] {
	result.Key = key
	return *result
}
func (result ItemsResult[T]) Items() (items []T, err error) {
	if result.IsError() {
		err = result.GetError()
		return
	}
	body := map[string][]T{}
	err = json.Unmarshal(result.Resp.Body(), &body)
	if err != nil {
		return
	}
	return body[result.Key], nil
}

type ItemResult[T any] struct {
	HttpResult
	Key string
}

func (result ItemResult[T]) Item() (item *T, err error) {
	if result.IsError() {
		item = nil
		err = fmt.Errorf("result is error: %v, %v", result.Err, result.Resp.StatusCode())
		return
	}
	if result.Key == "" {
		err = json.Unmarshal(result.Resp.Body(), &item)
		if err != nil {
			item = nil
		}
	} else {
		body := map[string]*T{}
		err := json.Unmarshal(result.Resp.Body(), &body)
		if err != nil {
			item = nil
		} else {
			item = body[result.Key]
		}
	}
	return
}

func NewItemsResult[T any](resp *session.Response, err error) *ItemsResult[T] {
	return &ItemsResult[T]{
		HttpResult: HttpResult{
			Resp: resp,
			Err:  err,
		},
	}
}
