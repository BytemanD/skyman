package utility

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
)

const (
	CODE_404 = 404
)

func CloneHeaders(h1 map[string]string, h2 map[string]string) map[string]string {
	clonedHeaders := map[string]string{}
	for k, v := range h1 {
		clonedHeaders[k] = v
	}
	for k, v := range h2 {
		clonedHeaders[k] = v
	}
	return clonedHeaders
}

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
func (resp Response) SaveBody(file *os.File, process bool) error {
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
	return err
}

func (resp Response) IsNotFound() bool {
	return resp.Status == 404
}
func EncodeHeaders(headers http.Header) string {
	headersString := []string{}
	for k, v := range getSafeHeaders(headers) {
		headersString = append(headersString, fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
	}
	return strings.Join(headersString, ", ")
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
func encodeHeaders(headers http.Header) string {
	headersString := []string{}
	for k, v := range getSafeHeaders(headers) {
		headersString = append(headersString, fmt.Sprintf("%s:%s", k, strings.Join(v, ",")))
	}
	return strings.Join(headersString, " ")
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
	session *http.Client
}

func (c RestfulClient) getClient() *http.Client {
	if c.session == nil {
		c.session = &http.Client{
			// Timeout: c.Timeout,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: c.Timeout,
				}).Dial,
			},
		}
	}
	return c.session
}

func (c RestfulClient) Request(req *http.Request) (*Response, error) {
	isIoStream := StringsContain(req.Header["Content-Type"], "application/octet-stream")
	encodedHeader := encodeHeaders(getSafeHeaders(req.Header))
	if isIoStream {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, encodedHeader, "<Omitted, octet-stream>")
	} else {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, encodedHeader, req.Body)
	}

	client := c.getClient()
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
		logging.Debug("RESP: [%d], \n    Body: %s", resp.StatusCode, "<Omitted, octet-stream>")
	} else {
		response.ReadAll()
		defer resp.Body.Close()
		logging.Debug("RESP: [%d], \n    Body: %s", response.Status, response.Body)
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
func ReadResponseBody(resp *http.Response) string {
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logging.Error("read reponse body failed: %s", err)
	}
	return string(bytes)
}

func UnmarshalResponse(resp *http.Response, o interface{}) error {
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, &o)
}

type Request struct {
	Url     string
	Body    []byte
	Query   url.Values
	Headers map[string]string
}
