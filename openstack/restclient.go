package openstack

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/utility"
)

type RestClient struct {
	BaseUrl    string
	Timeout    time.Duration
	session    *http.Client
	AuthPlugin auth.AuthPlugin
	Headers    map[string]string
}

func (c RestClient) getClient() *http.Client {
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

func (rest *RestClient) doRequest(method, reqUrl string, query url.Values,
	body []byte, headers map[string]string,
) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if query != nil {
		req.URL.RawQuery = query.Encode()
	}
	for k, v := range rest.Headers {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if rest.AuthPlugin != nil {
		rest.AuthPlugin.AuthRequest(req)
	}

	isIoStream := utility.StringsContain(req.Header["Content-Type"], "application/octet-stream")
	if isIoStream {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, utility.EncodeHeaders(req.Header), "<Omitted, octet-stream>")
	} else {
		logging.Debug("REQ: %s %s, \n    Headers: %v \n    Body: %v",
			req.Method, req.URL, utility.EncodeHeaders(req.Header), req.Body)
	}
	resp, err := rest.getClient().Do(req)
	if err != nil {
		return nil, err
	}
	response := utility.Response{
		Status:  resp.StatusCode,
		Reason:  resp.Status,
		Headers: resp.Header,
	}
	response.SetBodyReader(resp.Body)
	if isIoStream {
		logging.Debug("RESP: [%d], \n    Body: %s", response.Status, "<Omitted, octet-stream>")
	} else {
		response.ReadAll()
		logging.Debug("RESP: [%d], \n    Body: %s", response.Status, response.Body)
	}
	if resp.StatusCode >= 400 {
		return nil, &utility.HttpError{
			Status: resp.StatusCode, Reason: resp.Status,
			Message: response.BodyString(),
		}
	}
	return &response, nil
}
func (rest *RestClient) mustHasBaseUrl() error {
	if rest.BaseUrl == "" {
		return fmt.Errorf("base url is required")
	}
	return nil
}
func (rest *RestClient) versionUrl(url string) string {
	return utility.UrlJoin(rest.BaseUrl, url)
}

func (rest *RestClient) Index() (*utility.Response, error) {
	parsed, err := url.Parse(rest.BaseUrl)
	if err != nil {
		return nil, err
	}
	return rest.doRequest("GET",
		fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host),
		nil, nil, nil)
}

func (rest *RestClient) Get(url string, query url.Values) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.doRequest("GET", rest.versionUrl(url), query, nil, nil)
}
func (rest *RestClient) Post(url string, body []byte, query url.Values) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	if rest.BaseUrl == "" {
		return nil, fmt.Errorf("base url is required")
	}
	return rest.doRequest("POST", rest.versionUrl(url), query, body, nil)
}

func (rest *RestClient) Put(url string, body []byte, query url.Values) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.doRequest("PUT", rest.versionUrl(url), query, body, nil)
}
func (rest *RestClient) Delete(url string, query url.Values) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.doRequest("DELETE", rest.versionUrl(url), query, nil, nil)
}
func (rest *RestClient) Patch(req utility.Request) (*utility.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.doRequest("PATCH", rest.versionUrl(req.Url), req.Query, req.Body, req.Headers)
}
