package common

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

const (
	CODE_404 = 404
)

type BaseResponse interface {
	BodyString()
	GetHeader()
	BodyUnmarshal(object interface{})
}

type Response struct {
	Status     int
	Reason     string
	Body       []byte
	Headers    http.Header
	bodyReader io.ReadCloser
	readBody   bool
}

func (resp Response) GetHeader(key string) string {
	return resp.Headers.Get(key)
}
func (resp Response) GetContentLength() int {
	length, _ := strconv.Atoi(resp.GetHeader("Content-Length"))
	return length
}

func (resp *Response) BodyString() string {
	resp.ReadAll()
	return string(resp.Body)
}

func (resp *Response) ReadAll() error {
	if !resp.readBody {
		bodyBytes, err := io.ReadAll(resp.bodyReader)
		if err != nil {
			return err
		}
		defer resp.bodyReader.Close()
		resp.Body = bodyBytes
		resp.readBody = true
	}
	return nil
}

func (resp *Response) BodyUnmarshal(object interface{}) error {
	resp.ReadAll()
	return json.Unmarshal(resp.Body, object)
}
func (resp *Response) SaveBody(file *os.File, process bool) error {
	defer resp.bodyReader.Close()
	var reader io.Reader
	if process && resp.GetContentLength() > 0 {
		reader = &ReaderWithProcess{
			Reader: bufio.NewReaderSize(resp.bodyReader, 1024*32),
			Size:   resp.GetContentLength(),
		}
	} else {
		reader = bufio.NewReaderSize(resp.bodyReader, 1024*32)
	}
	_, err := io.Copy(bufio.NewWriter(file), reader)
	resp.readBody = true
	return err
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
	isIoStream := ContainsString(req.Header["Content-Type"], "application/octet-stream")

	if isIoStream {
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
	response := Response{
		bodyReader: resp.Body,
		Status:     resp.StatusCode,
		Reason:     resp.Status,
		Headers:    resp.Header,
	}
	if isIoStream {
		logging.Debug("Resp: status code: %d, body: %s", resp.StatusCode, "<Omitted, octet-stream>")
	} else {
		response.ReadAll()
		logging.Debug("Resp: status code: %d, body: %s", response.Status, response.Body)
	}
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
