package common

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type BaseResponse interface {
	BodyString()
	GetHeader()
	BodyUnmarshal()
}

type Response struct {
	Status  int
	Reason  string
	Body    []byte
	Headers http.Header
}

func (resp Response) BodyString() string {
	return string(resp.Body)
}

func (resp Response) GetHeader(key string) string {
	return resp.Headers.Get(key)
}

func (resp Response) BodyUnmarshal(object interface{}) error {
	return json.Unmarshal(resp.Body, object)
}

type Session struct {
}

func (session Session) Request(req *http.Request) (*Response, error) {
	logging.Debug("Req: %s %s with headers: %v, body: %v", req.Method, req.URL, req.Header, req.Body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	content, _ := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	logging.Debug("Resp: status code: %d, body: %s", resp.StatusCode, content)
	response := Response{
		Body:    content,
		Status:  resp.StatusCode,
		Reason:  resp.Status,
		Headers: resp.Header}
	return &response, response.JudgeStatus()
}

func (client Session) UrlJoin(path []string) string {
	return strings.Join(path, "/")
}

// func (session *Session) Get(url string, query netUrl.Values, headers map[string]string) (*Response, error) {
// 	req, err := http.NewRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	req.URL.RawQuery = query.Encode()
// 	return session.Request(req, headers)
// }

// func (session *Session) Post(url string, body []byte, headers map[string]string) (*Response, error) {
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return session.Request(req, headers)
// }

// func (session *Session) Put(url string, body []byte, headers map[string]string) (*Response, error) {
// 	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
// 	if err != nil {
// 		return nil, err
// 	}
// 	return session.Request(req, headers)
// }

// func (session *Session) Delete(url string, headers map[string]string) (*Response, error) {
// 	req, err := http.NewRequest("DELETE", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return session.Request(req, headers)
// }

type HttpError struct {
	Status  int
	Reason  string
	Message string
}

func (err *HttpError) Error() string {
	return err.Reason
}

func (resp *Response) JudgeStatus() error {
	switch {
	case resp.Status < 400:
		return nil
	default:
		return &HttpError{Status: resp.Status, Reason: resp.Reason,
			Message: resp.BodyString()}
	}
}
