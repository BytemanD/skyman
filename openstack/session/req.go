package session

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Request struct {
	*resty.Request
	Baseurl     string
	ResourceUrl string
}

func (r *Request) SetBody(body interface{}) *Request {
	r.Request.SetBody(body)
	return r
}
func (r *Request) ResetPath() *Request {
	r.ResourceUrl = ""
	return r
}
func (r *Request) SetQuery(query url.Values) *Request {
	r.Request.SetQueryParamsFromValues(query)
	return r
}
func (r *Request) SetHeader(key, value string) *Request {
	r.Request.SetHeader(key, value)
	return r
}
func (r *Request) SetResult(result interface{}) *Request {
	r.Request.SetResult(result)
	return r
}
func (r Request) buildUrl(path ...string) (string, error) {
	if !strings.HasPrefix(r.Baseurl, "http://") && !strings.HasPrefix(r.Baseurl, "https://") {
		return "", fmt.Errorf("invalid baseurl: %s", r.Baseurl)
	}
	paths := append([]string{r.ResourceUrl}, path...)
	url, err := url.JoinPath(r.Baseurl, paths...)
	if err != nil {
		return "", fmt.Errorf("invalid url( %s + %v): %w", r.Baseurl, strings.Join(path, "/"), err)
	}
	return url, nil
}
func (r Request) buildResponse(rawResp *resty.Response) (*Response, error) {
	resp := Response{rawResp}
	if resp.IsError() {
		return &resp, resp.Error()
	} else {
		return &resp, nil
	}
}

func (r Request) Get(path ...string) (*Response, error) {
	url, err := r.buildUrl(path...)
	if err != nil {
		return nil, err
	}
	rawResp, err := r.Request.Get(url)
	if err != nil {
		return nil, err
	}
	return r.buildResponse(rawResp)
}
func (r Request) Post(path ...string) (*Response, error) {
	url, err := r.buildUrl(path...)
	if err != nil {
		return nil, err
	}
	rawResp, err := r.Request.Post(url)
	if err != nil {
		return nil, err
	}
	return r.buildResponse(rawResp)
}
func (r Request) Put(path ...string) (*Response, error) {
	url, err := r.buildUrl(path...)
	if err != nil {
		return nil, err
	}
	rawResp, err := r.Request.Put(url)
	if err != nil {
		return nil, err
	}
	return r.buildResponse(rawResp)
}
func (r Request) Patch(path ...string) (*Response, error) {
	url, err := r.buildUrl(path...)
	if err != nil {
		return nil, err
	}
	rawResp, err := r.Request.Patch(url)
	if err != nil {
		return nil, err
	}
	return r.buildResponse(rawResp)
}

func (r Request) Delete(path ...string) (*Response, error) {
	url, err := r.buildUrl(path...)
	if err != nil {
		return nil, err
	}
	rawResp, err := r.Request.Delete(url)
	if err != nil {
		return nil, err
	}
	return r.buildResponse(rawResp)
}
