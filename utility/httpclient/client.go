package httpclient

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/easygo/pkg/stringutils"
)

type Client struct {
	Timeout     time.Duration
	AuthPlugin  AuthInterface
	BaseHeaders map[string]string
	session     *http.Client
}

func (c Client) getSession() *http.Client {
	if c.session == nil {
		c.session = &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: c.Timeout,
				}).Dial,
			},
		}
	}
	return c.session
}

func encodeHeaders(headers http.Header) string {
	headersString := []string{}
	for k, v := range headers {
		headersString = append(headersString, fmt.Sprintf("'%s: %s'", k, strings.Join(v, ",")))
	}
	return strings.Join(headersString, " ")
}

func (c Client) Request(req *http.Request) (*Response, error) {
	for key, value := range c.BaseHeaders {
		if len(req.Header.Values(key)) != 0 {
			continue
		}
		req.Header.Add(key, value)
	}
	if c.AuthPlugin != nil {
		if err := c.AuthPlugin.AuthRequest(req); err != nil {
			return nil, fmt.Errorf("auth request failed: %s", err)
		}
	}
	isIoStream := stringutils.ContainsString(req.Header.Values("Content-Type"), "application/octet-stream")
	encodedHeader := ""
	if c.AuthPlugin != nil {
		encodedHeader = encodeHeaders(c.AuthPlugin.GetSafeHeader(req.Header))
	} else {
		encodedHeader = encodeHeaders(req.Header)
	}
	if isIoStream {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, encodedHeader, "<Omitted, octet-stream>")
	} else {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, encodedHeader, req.Body)
	}
	session := c.getSession()
	resp, err := session.Do(req)

	if err != nil {
		return nil, err
	}
	response := Response{
		bodyReader: resp.Body,
		Status:     resp.StatusCode,
		Reason:     resp.Status,
		Headers:    resp.Header,
	}
	respBody := ""
	respIsIoStream := stringutils.ContainsString(resp.Header.Values("Content-Type"), "application/octet-stream")
	if respIsIoStream {
		respBody = "<Omitted, octet-stream>"
	} else {
		response.ReadAll()
		respBody = response.BodyString()
		defer resp.Body.Close()
	}
	logging.Debug("RESP: [%d], \n    Headers: %v\n    Body: %s",
		response.Status, encodeHeaders(resp.Header), respBody)
	return &response, response.MustNotError()
}
func (c Client) setHeader(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}
func (c Client) Get(url string, query url.Values, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	req.URL.RawQuery = query.Encode()
	return c.Request(req)
}
func (c Client) Post(url string, body []byte, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
func (c Client) Put(url string, body []byte, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
func (c Client) Delete(url string, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
func (c Client) Patch(url string, query url.Values, body []byte, headers map[string]string) (*Response, error) {
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	c.setHeader(req, headers)
	return c.Request(req)
}
