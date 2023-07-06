package openstack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	netUrl "net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type Session struct {
}

func (session *Session) Request(req *http.Request, headers map[string]string) (Response, error) {
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	logging.Debug("Req: %s %s with headers: %v, body: %v", req.Method, req.URL, headers, req.Body)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Response{}, err
	}
	content, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	response := Response{
		Body:    content,
		Status:  resp.StatusCode,
		Headers: resp.Header}
	logging.Debug("Status: %s, Body: %s", resp.Status, content)
	return response, response.JudgeStatus()
}

func (session *Session) Get(url string, query netUrl.Values, headers map[string]string) (Response, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.URL.RawQuery = query.Encode()
	return session.Request(req, headers)
}

func (session *Session) Post(url string, body []byte, headers map[string]string) (Response, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	return session.Request(req, headers)
}

func (session *Session) Put(url string, body []byte, headers map[string]string) (Response, error) {
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	return session.Request(req, headers)
}

func (session *Session) Delete(url string, headers map[string]string) (Response, error) {
	req, _ := http.NewRequest("DELETE", url, nil)
	return session.Request(req, headers)
}

type Response struct {
	Status  int
	Body    []byte
	Headers http.Header
}

type HttpError struct {
	Code   int
	Reason string
}

func (err *HttpError) Error() string {
	return fmt.Sprintf("%d %s", err.Code, err.Reason)
}

func (resp *Response) JudgeStatus() error {
	switch {
	case resp.Status <= 400:
		return nil
	case resp.Status == 400:
		return fmt.Errorf("BadRequest")
	case resp.Status == 404:
		return fmt.Errorf("%d NotFound", resp.Status)
	case resp.Status == 500:
		return fmt.Errorf("BadRequest")
	default:
		return fmt.Errorf("ErrorCode %d", resp.Status)
	}
}

func (resp *Response) BodyString() string {
	return string(resp.Body)
}

func (resp *Response) GetHeader(key string) string {
	return resp.Headers.Get(key)
}

func (resp *Response) BodyUnmarshal(object interface{}) error {
	return json.Unmarshal(resp.Body, object)
}
