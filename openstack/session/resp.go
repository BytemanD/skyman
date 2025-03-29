package session

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

const (
	CODE_404 = 404
)

var ErrHTTPStatus = errors.New("http error")
var ErrHTTP404 = fmt.Errorf("%w: %d", ErrHTTPStatus, CODE_404)

type Response struct{ *resty.Response }

func (r Response) RequestId() string {
	return r.RawResponse.Header.Get(HEADER_REQUEST_ID)
}

func (r Response) Error() error {
	if !r.IsError() {
		return nil
	}
	switch r.StatusCode() {
	case 404:
		return fmt.Errorf("%w: %s", ErrHTTP404, r.Body())
	default:
		return fmt.Errorf("%w: [%d], %s", ErrHTTPStatus, r.StatusCode(), string(r.Body()))
	}
}
func (r Response) IsNotFound() bool {
	return errors.Is(r.Error(), ErrHTTP404)
}
func (r Response) UnmarshalBody(v any) error {
	return json.Unmarshal(r.Body(), v)
}

func CheckResponse(resp *resty.Response) error {
	return Response{resp}.Error()
}
