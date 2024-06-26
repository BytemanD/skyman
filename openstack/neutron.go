package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/BytemanD/easygo/pkg/compare"
	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
)

type NeutronV2 struct {
	RestClient2
	currentVersion *model.ApiVersion
}
type RouterApi struct {
	NeutronV2
}
type NetworkApi struct {
	NeutronV2
}
type SubnetApi struct {
	NeutronV2
}

type PortApi struct {
	NeutronV2
}
type AgentApi struct {
	NeutronV2
}

func (c NeutronV2) Routers() RouterApi {
	return RouterApi{c}
}
func (c NeutronV2) Networks() NetworkApi {
	return NetworkApi{c}
}
func (c NeutronV2) Subnets() SubnetApi {
	return SubnetApi{c}
}
func (c NeutronV2) Ports() PortApi {
	return PortApi{c}
}
func (c NeutronV2) Agents() AgentApi {
	return AgentApi{c}
}
func (o *Openstack) getEndpoint() (string, error) {
	endpoint := os.Getenv("OS_NEUTRON_ENDPOINT")
	if endpoint != "" {
		return endpoint, nil
	}
	return o.AuthPlugin.GetServiceEndpoint("network", "neutron", "public")
}

func (o *Openstack) NeutronV2() *NeutronV2 {
	if o.neutronClient == nil {
		endpoint, err := o.getEndpoint()
		if err != nil {
			logging.Fatal("get neutron endpoint falied: %v", err)
		}
		o.neutronClient = &NeutronV2{
			RestClient2: NewRestClient2(utility.VersionUrl(endpoint, "v2.0"), o.AuthPlugin),
		}
	}
	return o.neutronClient
}
func (c *NeutronV2) GetCurrentVersion() (*model.ApiVersion, error) {
	if c.currentVersion == nil {
		resp, err := c.Index()
		if err != nil {
			return nil, err
		}
		apiVersions := struct{ Versions []model.ApiVersion }{}
		if err := resp.BodyUnmarshal(&apiVersions); err != nil {
			return nil, err
		}
		for _, version := range apiVersions.Versions {
			if version.Status == "CURRENT" {
				c.currentVersion = &version
			}
		}
	}
	if c.currentVersion != nil {
		return c.currentVersion, nil
	}
	return nil, fmt.Errorf("current version not found")
}

// router api

func (c RouterApi) List(query url.Values) ([]neutron.Router, error) {
	resp, err := c.NeutronV2.Get("routers", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]neutron.Router{"routers": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["routers"], nil
}
func (c RouterApi) ListByName(name string) ([]neutron.Router, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}

func (c RouterApi) Show(id string) (*neutron.Router, error) {
	resp, err := c.NeutronV2.Get(utility.UrlJoin("routers", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*neutron.Router{"router": {}}
	resp.BodyUnmarshal(&body)
	return body["router"], err
}
func (c RouterApi) Create(params map[string]interface{}) (*neutron.Router, error) {
	data := map[string]interface{}{"router": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := c.NeutronV2.Post("routers", body, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*neutron.Router{"router": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["router"], err
}

func (c RouterApi) Delete(id string) (err error) {
	_, err = c.NeutronV2.Delete(utility.UrlJoin("routers", id), nil)
	return err
}

// subnet api

func (c NetworkApi) List(query url.Values) ([]neutron.Network, error) {
	resp, err := c.NeutronV2.Get("networks", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]neutron.Network{"networks": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["networks"], nil
}
func (c NetworkApi) ListByName(name string) ([]neutron.Network, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}

func (c NetworkApi) Show(id string) (*neutron.Network, error) {
	resp, err := c.NeutronV2.Get(utility.UrlJoin("networks", id), nil)
	if err != nil {
		return nil, err
	}
	body := struct {
		Network neutron.Network `json:"network,omitempty"`
	}{}
	resp.BodyUnmarshal(&body)
	return &body.Network, err
}
func (c NetworkApi) Found(idOrName string) (*neutron.Network, error) {
	network, err := c.Show(idOrName)
	if err == nil {
		return network, err
	}
	networks, err := c.ListByName(idOrName)
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	if len(networks) == 1 {
		return &(networks[0]), nil
	}
	if len(networks) == 0 {
		return nil, fmt.Errorf("network %s not found", idOrName)
	} else {
		return nil, fmt.Errorf("found multi networks named %s", idOrName)
	}
}
func (c NetworkApi) Create(params map[string]interface{}) (*neutron.Network, error) {
	data := map[string]interface{}{"network": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := c.NeutronV2.Post("networks", body, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*neutron.Network{"network": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["network"], err
}

func (c NetworkApi) Delete(id string) (err error) {
	_, err = c.NeutronV2.Delete(utility.UrlJoin("networks", id), nil)
	return err
}

// subnet api

func (c SubnetApi) List(query url.Values) ([]neutron.Subnet, error) {
	resp, err := c.NeutronV2.Get("subnets", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]neutron.Subnet{"subnets": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["subnets"], nil
}
func (c SubnetApi) ListByName(name string) ([]neutron.Subnet, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}

func (c SubnetApi) Show(id string) (*neutron.Subnet, error) {
	resp, err := c.NeutronV2.Get(utility.UrlJoin("subnets", id), nil)
	if err != nil {
		return nil, err
	}
	body := struct {
		Subnet neutron.Subnet `json:"subnet,omitempty"`
	}{}

	resp.BodyUnmarshal(&body)
	return &body.Subnet, err
}
func (c SubnetApi) Found(idOrName string) (*neutron.Subnet, error) {
	subnet, err := c.Show(idOrName)
	if err == nil {
		return subnet, err
	}
	subnets, err := c.ListByName(idOrName)
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	if len(subnets) == 1 {
		return &(subnets[0]), nil
	}
	if len(subnets) == 0 {
		return nil, fmt.Errorf("subnet %s not found", idOrName)
	} else {
		return nil, fmt.Errorf("found multi subnets named %s", idOrName)
	}
}
func (c SubnetApi) Create(params map[string]interface{}) (*neutron.Subnet, error) {
	data := map[string]interface{}{"subnet": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := c.NeutronV2.Post("subnets", body, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*neutron.Subnet{"subnet": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["subnet"], err
}

func (c SubnetApi) Delete(id string) (err error) {
	_, err = c.NeutronV2.Delete(utility.UrlJoin("subnets", id), nil)
	return err
}

// port api

func (c PortApi) List(query url.Values) ([]neutron.Port, error) {
	resp, err := c.NeutronV2.Get("ports", query)
	if err != nil {
		return nil, err
	}
	body := struct {
		Ports []neutron.Port `json:"ports"`
	}{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Ports, nil
}
func (c PortApi) ListByName(name string) ([]neutron.Port, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}

func (c PortApi) Show(id string) (*neutron.Port, error) {
	resp, err := c.NeutronV2.Get(utility.UrlJoin("ports", id), nil)
	if err != nil {
		return nil, err
	}
	body := struct {
		Port *neutron.Port `json:"port"`
	}{}
	// body := map[string]*neutron.Port{"ports": {}}
	resp.BodyUnmarshal(&body)
	return body.Port, err
}
func (c PortApi) Found(idOrName string) (*neutron.Port, error) {
	subnet, err := c.Show(idOrName)
	if err == nil {
		return subnet, err
	}
	ports, err := c.ListByName(idOrName)
	if compare.IsType[httpclient.HttpError](err) {
		httpError, _ := err.(httpclient.HttpError)
		if !httpError.IsNotFound() {
			return nil, err
		}
	}
	if len(ports) == 1 {
		return &(ports[0]), nil
	}
	if len(ports) == 0 {
		return nil, fmt.Errorf("port %s not found", idOrName)
	} else {
		return nil, fmt.Errorf("found multi ports named %s", idOrName)
	}
}
func (c PortApi) Create(params map[string]interface{}) (*neutron.Port, error) {
	data := map[string]interface{}{"port": params}
	body, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	resp, err := c.NeutronV2.Post("ports", body, nil)
	if err != nil {
		return nil, err
	}
	respBody := map[string]*neutron.Port{"subnet": {}}
	err = resp.BodyUnmarshal(&respBody)
	return respBody["port"], err
}

func (c PortApi) Delete(id string) (err error) {
	_, err = c.NeutronV2.Delete(utility.UrlJoin("ports", id), nil)
	return err
}
func (c AgentApi) List(query url.Values) ([]neutron.Agent, error) {
	resp, err := c.NeutronV2.Get("agents", query)
	if err != nil {
		return nil, err
	}
	body := struct {
		Agents []neutron.Agent `json:"agents"`
	}{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Agents, nil
}
