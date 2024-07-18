package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type Response struct {
	Status     int
	Reason     string
	Body       []byte
	Headers    http.Header
	bodyReader io.ReadCloser
}

func (resp *Response) SetBodyReader(reader io.ReadCloser) {
	resp.bodyReader = reader
}

func (resp Response) BodyString() string {
	return string(resp.Body)
}

func (resp *Response) ReadAll() error {
	var err error
	resp.Body, err = io.ReadAll(resp.bodyReader)
	defer resp.bodyReader.Close()
	return err
}

func (resp Response) GetHeader(key string) string {
	return resp.Headers.Get(key)
}

func (resp Response) BodyUnmarshal(object interface{}) error {
	return json.Unmarshal(resp.Body, object)
}
func (resp Response) GetContentLength() int {
	length, _ := strconv.Atoi(resp.Headers.Get("Content-Length"))
	return length
}

func (resp Response) IsNotFound() bool {
	return resp.Status == 404
}
func (resp *Response) MustNotError() error {
	switch {
	case resp.Status < 400:
		return nil
	default:
		return HttpError{Status: resp.Status, Reason: resp.Reason,
			Message: resp.BodyString()}
	}
}
func (resp *Response) BodyReader() io.ReadCloser {
	return resp.bodyReader
}

func (resp *Response) CloseReader() error {
	return resp.bodyReader.Close()
}

func MustNotError(resp *resty.Response) error {
	if resp.IsError() {
		return HttpError{
			Status: resp.StatusCode(), Reason: resp.Status(),
			Message: string(resp.Body()),
		}
	}
	return nil
}
