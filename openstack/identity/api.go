package identity

import (
	"net/url"
)

func (client IdentityClientV3) ServiceList(query url.Values) ([]Service, error) {
	body := map[string][]Service{"services": {}}
	err := client.List("services", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["services"], nil
}
func (client IdentityClientV3) ServiceListByName(name string) ([]Service, error) {
	query := url.Values{}
	query.Add("name", name)
	return client.ServiceList(query)
}
func (client IdentityClientV3) ServiceShow(serviceId string) (*Service, error) {
	body := map[string]*Service{"service": {}}
	err := client.Show("services", serviceId, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["service"], nil
}
func (client IdentityClientV3) EndpointList(query url.Values) ([]Endpoint, error) {
	body := map[string][]Endpoint{"endpoints": {}}
	err := client.List("endpoints", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["endpoints"], nil
}
