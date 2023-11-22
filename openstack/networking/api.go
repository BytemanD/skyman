package networking

import (
	"net/url"

	"github.com/BytemanD/skyman/openstack/common"
)

// router api
func (client NeutronClientV2) RouterList(query url.Values) ([]Router, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "routers", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Router{"routers": {}}
	resp.BodyUnmarshal(&body)
	return body["routers"], nil
}
func (client NeutronClientV2) RouterListByName(name string) ([]Router, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.RouterList(query)
}
func (client NeutronClientV2) RouterShow(id string) (*Router, error) {
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "routers", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Router{"router": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["router"], nil
}
func (client NeutronClientV2) RouterDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "routers", id, client.BaseHeaders),
	)
	return err
}

// network api
func (client NeutronClientV2) NetworkList(query url.Values) ([]Network, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "networks", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Network{"networks": {}}
	resp.BodyUnmarshal(&body)
	return body["networks"], nil
}
func (client NeutronClientV2) NetworkListByName(name string) ([]Network, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.NetworkList(query)
}
func (client NeutronClientV2) NetworkShow(id string) (*Network, error) {
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "networks", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Network{"network": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["network"], nil
}
func (client NeutronClientV2) NetworkDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "networks", id, client.BaseHeaders),
	)
	return err
}

// port api
func (client NeutronClientV2) PortList(query url.Values) ([]Port, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "ports", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Port{"ports": {}}
	resp.BodyUnmarshal(&body)
	return body["ports"], nil
}
func (client NeutronClientV2) PortListByName(name string) ([]Port, error) {
	query := url.Values{}
	query.Set("name", name)
	return client.PortList(query)
}
func (client NeutronClientV2) PortShow(id string) (*Port, error) {
	resp, err := client.Request(
		common.NewResourceShowRequest(client.endpoint, "ports", id, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Port{"port": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["port"], nil
}
func (client NeutronClientV2) PortDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "ports", id, client.BaseHeaders),
	)
	return err
}
