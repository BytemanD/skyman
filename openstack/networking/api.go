package networking

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/common"
	"github.com/BytemanD/skyman/utility"
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
func (client NeutronClientV2) RouterCreate(params map[string]interface{}) (*Router, error) {
	data := map[string]interface{}{"router": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "routers", body, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Router{"router": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["router"], err
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
func (client NeutronClientV2) NetworkFound(idOrName string) (*Network, error) {
	network, err := client.NetworkShow(idOrName)
	if err == nil {
		return network, nil
	}
	if httpError, ok := err.(*utility.HttpError); ok {
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	networks, err := client.NetworkListByName(idOrName)
	if err != nil {
		return nil, err
	}
	if len(networks) == 1 {
		return &(networks[0]), nil
	}
	if len(networks) == 0 {
		return nil, fmt.Errorf("network named %s not found", idOrName)
	} else {
		return nil, fmt.Errorf("found multi networks named %s", idOrName)
	}
}

func (client NeutronClientV2) NetworkDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "networks", id, client.BaseHeaders),
	)
	return err
}
func (client NeutronClientV2) NetworkCreate(params map[string]interface{}) (*Network, error) {
	data := map[string]interface{}{"network": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "networks", body, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Network{"network": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["network"], err
}

// subnet api
func (client NeutronClientV2) SubnetList(query url.Values) ([]Subnet, error) {
	resp, err := client.Request(
		common.NewResourceListRequest(client.endpoint, "subnets", query, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	body := map[string][]Subnet{"subnets": {}}
	resp.BodyUnmarshal(&body)
	return body["subnets"], nil
}
func (client NeutronClientV2) SubnetCreate(params map[string]interface{}) (*Subnet, error) {
	data := map[string]interface{}{"subnet": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := client.Request(
		common.NewResourceCreateRequest(client.endpoint, "subnets", body, client.BaseHeaders),
	)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Subnet{"subnet": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["subnet"], err
}
func (client NeutronClientV2) SubnetDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint, "subnets", id, client.BaseHeaders),
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
