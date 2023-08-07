package networking

import (
	"net/url"
)

// router api
func (client NeutronClientV2) RouterList(query url.Values) []Router {
	body := map[string][]Router{"routers": {}}
	client.List("routers", query, nil, &body)
	return body["routers"]
}
func (client NeutronClientV2) RouterListByName(name string) []Router {
	query := url.Values{}
	query.Set("name", name)
	return client.RouterList(query)
}
func (client NeutronClientV2) RouterShow(id string) (*Router, error) {
	router := Router{}
	err := client.Show("routers", id, client.BaseHeaders, &router)
	if err != nil {
		return nil, err
	}
	return &router, nil
}

// network api
func (client NeutronClientV2) NetworkList(query url.Values) ([]Network, error) {
	body := map[string][]Network{"routers": {}}
	err := client.List("networks", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["networks"], nil
}
func (client NeutronClientV2) NetworkListByName(name string) ([]Network, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.NetworkList(query)
}
func (client NeutronClientV2) NetworkShow(id string) (*Network, error) {
	network := Network{}
	err := client.Show("networks", id, client.BaseHeaders, &network)
	if err != nil {
		return nil, err
	}
	return &network, nil
}

// port api
func (client NeutronClientV2) PortList(query url.Values) []Port {
	body := map[string][]Port{"ports": {}}
	client.List("ports", query, nil, &body)
	return body["ports"]
}
func (client NeutronClientV2) PortListByName(name string) []Port {
	query := url.Values{}
	query.Set("name", name)
	return client.PortList(query)
}
func (client NeutronClientV2) PortShow(id string) (*Port, error) {
	port := Port{}
	err := client.Show("ports", id, client.BaseHeaders, &port)
	if err != nil {
		return nil, err
	}
	return &port, nil
}
