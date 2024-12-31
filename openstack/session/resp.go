package session

import (
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type Response struct{ *resty.Response }

func (r Response) RequestId() string {
	return r.RawResponse.Header.Get(HEADER_REQUEST_ID)
}

func (r Response) Error() error {
	return HttpError{
		Status:  r.StatusCode(),
		Reason:  r.Status(),
		Message: string(r.Body())}
}
func (r Response) IsNotFound() bool {
	return r.StatusCode() == CODE_404
}
func (r Response) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(r.Body(), v)
}
