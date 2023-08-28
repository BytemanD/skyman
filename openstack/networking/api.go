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
	resp := map[string]*Router{"router": {}}
	err := client.Show("routers", id, client.BaseHeaders, &resp)
	if err != nil {
		return nil, err
	}
	return resp["router"], nil
}
func (client NeutronClientV2) RouterDelete(id string) error {
	err := client.Delete("routers", id, client.BaseHeaders)
	if err != nil {
		return err
	}
	return nil
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
	resp := map[string]*Network{"network": {}}
	err := client.Show("networks", id, client.BaseHeaders, &resp)
	if err != nil {
		return nil, err
	}
	return resp["network"], nil
}
func (client NeutronClientV2) NetworkDelete(id string) error {
	err := client.Delete("networks", id, client.BaseHeaders)
	if err != nil {
		return err
	}
	return nil
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
	resp := map[string]*Port{"port": {}}
	err := client.Show("ports", id, client.BaseHeaders, &resp)
	if err != nil {
		return nil, err
	}
	return resp["port"], nil
}
func (client NeutronClientV2) PortDelete(id string) error {
	err := client.Delete("ports", id, client.BaseHeaders)
	if err != nil {
		return err
	}
	return nil
}
