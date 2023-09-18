package identity

import (
	"bytes"
	"encoding/json"
	"fmt"
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
func getValue(respObj map[string]interface{}, key string) (interface{}, error) {
	if key == "" {
		return respObj, nil
	} else if obj, ok := respObj[key]; ok {
		return obj, nil
	} else {
		return nil, fmt.Errorf("key %s not found in body", key)
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
func (client RestfuleClient) FoundByIdOrName(resource string, idOrName string,
	key string, pluralKey string, headers map[string]string,
) (interface{}, error) {
	respObj := map[string]interface{}{}
	err := client.Show(resource, idOrName, headers, &respObj)
	if err == nil {
		return getValue(respObj, key)
	}
	query := url.Values{}
	query.Set("name", idOrName)
	respObjs := map[string][]map[string]string{}
	err = client.List(resource, query, headers, &respObjs)
	if err == nil {
		if objs, ok := respObjs[pluralKey]; ok {
			if len(objs) == 1 {
				if err := client.Show(resource, objs[0]["id"], headers, &respObj); err != nil {
					return nil, err
				}
				return getValue(respObj, key)
			} else if len(objs) == 0 {
				return nil, fmt.Errorf("%s %s not found", key, idOrName)
			} else {
				return nil, fmt.Errorf("found multi %s %s", resource, idOrName)
			}
		}
	}
	return nil, err
}

func (client RestfuleClient) Put(resource string, body []byte, headers map[string]string, obj interface{},
) error {
	req, err := http.NewRequest("PUT", client.getUrl(resource), bytes.NewBuffer(body))
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
	req, err := http.NewRequest("POST", client.getUrl(resource), bytes.NewBuffer(body))
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
		// return fmt.Errorf("%s", resp.BodyString())
	}
	if obj != nil {
		json.Unmarshal(resp.Body, &obj)
	}
	return nil
}
