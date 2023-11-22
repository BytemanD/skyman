package common

import (
	"fmt"
	"net/url"
)

type RestfulRequest struct {
	Endpoint string
	Method   string
	Resource string
	Id       string
	Query    url.Values
	Body     []byte
	Headers  map[string]string
}

func (request *RestfulRequest) Url() (string, error) {
	elem := []string{}
	if request.Resource != "" {
		elem = append(elem, request.Resource)
	}
	if request.Id != "" {
		elem = append(elem, request.Id)
	}
	return url.JoinPath(request.Endpoint, elem...)
}

func NewIndexRequest(endpoint string, query url.Values, headers map[string]string,
) RestfulRequest {
	baseUrl, _ := url.Parse(endpoint)
	indexUrl := fmt.Sprintf(
		"%s://%s:%s/", baseUrl.Scheme, baseUrl.Hostname(), baseUrl.Port(),
	)
	return RestfulRequest{
		Method: "GET", Endpoint: indexUrl,
		Query: query, Headers: headers,
	}
}

func NewResourceListRequest(endpoint string, resource string, query url.Values,
	headers map[string]string,
) RestfulRequest {
	return RestfulRequest{
		Method: "GET", Endpoint: endpoint, Resource: resource,
		Query: query, Headers: headers,
	}
}

func NewResourceShowRequest(endpoint string, resource string, id string,
	headers map[string]string,
) RestfulRequest {
	return RestfulRequest{
		Method: "GET", Endpoint: endpoint, Resource: resource,
		Id:      id,
		Headers: headers,
	}
}

func NewResourceCreateRequest(endpoint string, resource string, body []byte,
	headers map[string]string,
) RestfulRequest {
	return RestfulRequest{
		Method: "POST", Endpoint: endpoint, Resource: resource,
		Body: body, Headers: headers,
	}
}

func NewResourcePutRequest(endpoint string, resource string, id string, body []byte,
	headers map[string]string,
) RestfulRequest {
	return RestfulRequest{
		Method: "PUT", Endpoint: endpoint, Resource: resource,
		Id: id, Body: body, Headers: headers,
	}
}
func NewResourceDeleteRequest(endpoint string, resource string, id string,
	headers map[string]string,
) RestfulRequest {
	return RestfulRequest{
		Method: "DELETE", Endpoint: endpoint, Resource: resource,
		Id: id, Headers: headers,
	}
}
