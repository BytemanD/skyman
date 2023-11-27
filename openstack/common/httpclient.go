package common

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

type BaseResponse interface {
	BodyString()
	GetHeader()
	BodyUnmarshal(object interface{})
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
func (resp Response) IsNotFound() bool {
	return resp.Status == 404
}

type Session struct {
	Timeout time.Duration
}

func getSafeHeaders(headers http.Header) http.Header {
	safeHeaders := map[string][]string{}
	for k, v := range headers {
		if k == "X-Auth-Token" {
			safeHeaders[k] = []string{"******"}
		} else {
			safeHeaders[k] = v
		}
	}
	return safeHeaders
}

func (session Session) Request(req *http.Request) (*Response, error) {
	logging.Debug("Req: %s %s with headers: %v, body: %v", req.Method, req.URL,
		getSafeHeaders(req.Header), req.Body)

	client := http.Client{Timeout: session.Timeout}
	resp, err := client.Do(req)

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

const (
	CODE_404 = 404
)

type HttpError struct {
	Status  int
	Reason  string
	Message string
}

func (err HttpError) Error() string {
	return err.Reason
}
func (err HttpError) IsNotFound() bool {
	return err.Status == CODE_404
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

type RestfulClient struct {
	Timeout time.Duration
}

func (c RestfulClient) Request(req *http.Request) (*Response, error) {
	if ContainsString(req.Header["Content-Type"], "application/octet-stream") {
		logging.Debug("Req: %s %s with headers: %v, body: %v", req.Method, req.URL,
			getSafeHeaders(req.Header), "<Omitted, octet-stream>")
	} else {
		logging.Debug("Req: %s %s with headers: %v, body: %v", req.Method, req.URL,
			getSafeHeaders(req.Header), req.Body)
	}

	client := http.Client{Timeout: c.Timeout}
	resp, err := client.Do(req)

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
func (c RestfulClient) setHeader(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}

func (c RestfulClient) Get(url string, query url.Values, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	req.URL.RawQuery = query.Encode()
	return c.Request(req)
}
func (c RestfulClient) Post(url string, body []byte, headers map[string]string,
) (*Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
func (c RestfulClient) Put(url string, body []byte, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
func (c RestfulClient) Delete(url string, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest("Delete", url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
