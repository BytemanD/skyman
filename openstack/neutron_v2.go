package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
)

const (
	URL_LIST_SECURITY_GROUPS      = "security-groups"
	URL_SHOW_SECURITY_GROUP       = "security-groups/%s"
	URL_LIST_SECURITY_GROUP_RULES = "security-group-rules"
	URL_SHOW_SECURITY_GROUP_RULE  = "security-group-rules/%s"
)

type NeutronV2 struct {
	RestClient2
	currentVersion *model.ApiVersion
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
		body := struct{ Versions []model.ApiVersion }{}
		if err := json.Unmarshal(resp.Body(), &body); err != nil {
			return nil, err
		}
		for _, version := range body.Versions {
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
type routerApi struct {
	ResourceApi
}

func (c NeutronV2) Router() routerApi {
	return routerApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "routers",
			Client:      c.session,
			SingularKey: "router",
			PluralKey:   "routers",
		},
	}
}
func (c routerApi) List(query url.Values) ([]neutron.Router, error) {
	body := struct{ Routers []neutron.Router }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Routers, nil
}
func (c routerApi) ListByName(name string) ([]neutron.Router, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c routerApi) Show(id string) (*neutron.Router, error) {
	body := struct{ Router neutron.Router }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Router, nil

}
func (c routerApi) Create(params map[string]interface{}) (*neutron.Router, error) {
	body := struct{ Router neutron.Router }{}
	if _, err := c.SetBody(map[string]interface{}{"router": params}).Post(&body); err != nil {
		return nil, err
	}
	return &body.Router, nil
}

func (c routerApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// network api
type networkApi struct {
	ResourceApi
}

func (c NeutronV2) Network() networkApi {
	return networkApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "networks",
			Client:      c.session,
			SingularKey: "network",
			PluralKey:   "networks",
		},
	}
}

func (c networkApi) List(query url.Values) ([]neutron.Network, error) {
	body := struct{ Networks []neutron.Network }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Networks, nil
}
func (c networkApi) ListByName(name string) ([]neutron.Network, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c networkApi) Show(id string) (*neutron.Network, error) {
	body := struct{ Network neutron.Network }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Network, nil
}
func (c networkApi) Found(idOrName string) (*neutron.Network, error) {
	return FoundResource[neutron.Network](c.ResourceApi, idOrName)
}
func (c networkApi) Create(params map[string]interface{}) (*neutron.Network, error) {
	body := struct{ Network neutron.Network }{}
	if _, err := c.SetBody(map[string]interface{}{"network": params}).Post(&body); err != nil {
		return nil, err
	}
	return &body.Network, nil
}

func (c networkApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// subnet api
type subnetApi struct {
	ResourceApi
}

func (c NeutronV2) Subnet() subnetApi {
	return subnetApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "subnets",
			Client:      c.session,
			SingularKey: "subnet",
			PluralKey:   "subnets",
		},
	}
}
func (c subnetApi) List(query url.Values) ([]neutron.Subnet, error) {
	body := struct{ Subnets []neutron.Subnet }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Subnets, nil
}
func (c subnetApi) ListByName(name string) ([]neutron.Subnet, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c subnetApi) Show(id string) (*neutron.Subnet, error) {
	body := struct{ Subnet neutron.Subnet }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Subnet, nil
}
func (c subnetApi) Found(idOrName string) (*neutron.Subnet, error) {
	return FoundResource[neutron.Subnet](c.ResourceApi, idOrName)
}
func (c subnetApi) Create(params map[string]interface{}) (*neutron.Subnet, error) {
	body := struct{ Subnet neutron.Subnet }{}
	if _, err := c.SetBody(map[string]interface{}{"subnet": params}).Post(&body); err != nil {
		return nil, err
	}
	return &body.Subnet, nil
}

func (c subnetApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// port api
type portApi struct{ ResourceApi }

func (c NeutronV2) Port() portApi {
	return portApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "ports",
			Client:      c.session,
			SingularKey: "port",
			PluralKey:   "ports",
		},
	}
}

func (c portApi) List(query url.Values) ([]neutron.Port, error) {
	body := struct{ Ports []neutron.Port }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Ports, nil
}
func (c portApi) ListByName(name string) ([]neutron.Port, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}

func (c portApi) Show(id string) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Port, nil
}
func (c portApi) Found(idOrName string) (*neutron.Port, error) {
	return FoundResource[neutron.Port](c.ResourceApi, idOrName)
}
func (c portApi) Create(params map[string]interface{}) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.SetBody(map[string]interface{}{"port": params}).Post(&body); err != nil {
		return nil, err
	}
	return &body.Port, nil
}

func (c portApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// neutron agent api
type agentApi struct {
	ResourceApi
}

func (c NeutronV2) Agent() agentApi {
	return agentApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "agents",
			Client:      c.session,
			SingularKey: "agent",
			PluralKey:   "agents",
		},
	}
}

func (c agentApi) List(query url.Values) ([]neutron.Agent, error) {
	body := struct{ Agents []neutron.Agent }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Agents, nil
}

// security group api

type sgApi struct {
	ResourceApi
}

func (c NeutronV2) SecurityGroup() sgApi {
	return sgApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "security-groups",
			Client:      c.session,
			SingularKey: "security_group",
			PluralKey:   "security_groups",
		},
	}
}

func (c sgApi) List(query url.Values) ([]neutron.SecurityGroup, error) {
	body := struct {
		SecurityGroups []neutron.SecurityGroup `json:"security_groups"`
	}{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.SecurityGroups, nil
}
func (c sgApi) Show(id string) (*neutron.SecurityGroup, error) {
	body := struct {
		SecurityGroup neutron.SecurityGroup `json:"security_group"`
	}{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.SecurityGroup, nil
}
func (c sgApi) Found(idOrName string) (*neutron.SecurityGroup, error) {
	return FoundResource[neutron.SecurityGroup](c.ResourceApi, idOrName)
}

// security group rule api
type sgRuleApi struct {
	ResourceApi
}

func (c NeutronV2) SecurityGroupRule() sgRuleApi {
	return sgRuleApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "security-group-rules",
			Client:      c.session,
			SingularKey: "security_group_rule",
			PluralKey:   "security_group_rules",
		},
	}
}

func (c sgRuleApi) List(query url.Values) ([]neutron.SecurityGroupRule, error) {
	body := struct {
		SecurityGroupRules []neutron.SecurityGroupRule `json:"security_group_rules"`
	}{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.SecurityGroupRules, nil
}

func (c sgRuleApi) Show(id string) (*neutron.SecurityGroupRule, error) {
	body := struct {
		SecurityGroupRules neutron.SecurityGroupRule `json:"security_group_rule"`
	}{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.SecurityGroupRules, nil
}

// qos policy api
type qosPolicyApi struct {
	ResourceApi
}

func (c NeutronV2) QosPolicy() qosPolicyApi {
	return qosPolicyApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "qos/policies",
			Client:      c.session,
			SingularKey: "policies",
			PluralKey:   "policies",
		},
	}
}
func (c qosPolicyApi) List(query url.Values) ([]neutron.QosPolicy, error) {
	body := struct{ Policies []neutron.QosPolicy }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	return body.Policies, nil
}
func (c qosPolicyApi) Show(id string) (*neutron.QosPolicy, error) {
	body := struct{ Policy neutron.QosPolicy }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Policy, nil
}

func (c qosPolicyApi) Found(idOrName string) (*neutron.QosPolicy, error) {
	return FoundResource[neutron.QosPolicy](c.ResourceApi, idOrName)
}

// qos rule api
type qosRuleApi struct {
	ResourceApi
}

func (c NeutronV2) QosRule() qosRuleApi {
	return qosRuleApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "qos/policies",
			Client:      c.session,
			SingularKey: "policies",
			PluralKey:   "policies",
		},
	}
}

// func (c qosPolicyApi) List(query url.Values) ([]neutron.QosPolicy, error) {
// 	body := struct{ Policies []neutron.QosPolicy }{}
// 	if _, err := c.SetQuery(query).Get(&body); err != nil {
// 		return nil, err
// 	}
// 	return body.Policies, nil
// }
