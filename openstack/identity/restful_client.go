package identity

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type RestfuleClient struct {
	V3AuthClient
	Endpoint string
}

func updateHeaders(req *http.Request, headers map[string]string) {
	for k, v := range headers {
		req.Header.Set(k, v)
	}
}
func (client RestfuleClient) Index(obj interface{}) error {
	req, err := http.NewRequest("GET", client.getUrl(), nil)
	if err != nil {
		return err
	}
	return client.doRequest(req, obj)
}
func (client RestfuleClient) List(resource string, query url.Values,
	headers map[string]string, obj interface{},
) error {
	req, err := http.NewRequest("GET", client.getUrl(resource), nil)
	if err != nil {
		return err
	}
	req.URL.RawQuery = query.Encode()
	updateHeaders(req, headers)
	return client.doRequest(req, obj)
}

func (client RestfuleClient) Show(resource string, id string, headers map[string]string,
	obj interface{},
) error {
	req, err := http.NewRequest("GET", client.getUrl(resource, id), nil)
	if err != nil {
		return err
	}
	updateHeaders(req, headers)
	return client.doRequest(req, obj)
}
func (client RestfuleClient) Delete(resource string, id string, headers map[string]string) error {
	req, err := http.NewRequest("DELETE", client.getUrl(resource, id), nil)
	if err != nil {
		return err
	}
	updateHeaders(req, headers)
	return client.doRequest(req, nil)
}

func (client RestfuleClient) Create(resource string, body []byte, headers map[string]string,
	obj interface{},
) error {
	req, err := http.NewRequest("DELETE", client.getUrl(resource), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	updateHeaders(req, headers)
	return client.doRequest(req, obj)
}

func (client RestfuleClient) getUrl(path ...string) string {
	return client.session.UrlJoin(append([]string{client.Endpoint}, path...))
}

func (client RestfuleClient) doRequest(req *http.Request, obj interface{}) error {
	resp, err := client.Request(req)
	if err != nil {
		return err
	}
	if obj != nil {
		json.Unmarshal(resp.Body, &obj)
	}
	return nil
}
