package internal

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/utility"
)

const (
	ROUTERS                       = "routers"
	NETWORKS                      = "networks"
	SUBNETS                       = "subnets"
	PORTS                         = "ports"
	URL_LIST_SECURITY_GROUPS      = "security-groups"
	URL_SHOW_SECURITY_GROUP       = "security-groups/%s"
	URL_LIST_SECURITY_GROUP_RULES = "security-group-rules"
	URL_SHOW_SECURITY_GROUP_RULE  = "security-group-rules/%s"
)

type NeutronV2 struct {
	*ServiceClient
	currentVersion *model.ApiVersion
}

func (c *NeutronV2) GetCurrentVersion() (*model.ApiVersion, error) {
	if c.currentVersion == nil {
		body := struct{ Versions []model.ApiVersion }{}
		_, err := c.Index(&body)
		if err != nil {
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
type routerApi struct{ ResourceApi }

func (c NeutronV2) Router() routerApi {
	return routerApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
	}
}
func (c routerApi) List(query url.Values) ([]neutron.Router, error) {
	body := struct{ Routers []neutron.Router }{}
	if _, err := c.Get(ROUTERS, query, &body); err != nil {
		return nil, err
	}
	return body.Routers, nil
}
func (c routerApi) ListByName(name string) ([]neutron.Router, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c routerApi) Show(id string) (*neutron.Router, error) {
	body := struct{ Router neutron.Router }{}
	if _, err := c.Get("routers/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.Router, nil

}
func (c routerApi) Create(params map[string]interface{}) (*neutron.Router, error) {
	body := struct{ Router neutron.Router }{}
	if _, err := c.Post(ROUTERS, map[string]interface{}{"router": params}, &body); err != nil {
		return nil, err
	}
	return &body.Router, nil
}

func (c routerApi) Delete(id string) error {
	_, err := c.ResourceApi.Delete("routers/" + id)
	return err
}
func (c routerApi) Found(idOrName string) (*neutron.Router, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// Interface: subnet-id | port=port-id
func (c routerApi) AddSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.Put(
		fmt.Sprintf("routers/%s/add_router_interface", routerId),
		body, nil,
	); err != nil {
		return err
	}
	return nil
}
func (c routerApi) AddPort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.Put(
		fmt.Sprintf("routers/%s/add_router_interface", routerId),
		body, nil,
	); err != nil {
		return err
	}
	return nil
}
func (c routerApi) RemoveSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.Put(
		fmt.Sprintf("routers/%s/remove_router_interface", routerId),
		body, nil,
	); err != nil {
		return err
	}
	return nil
}
func (c routerApi) RemovePort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.Put(
		fmt.Sprintf("routers/%s/remove_router_interface", routerId),
		body, nil,
	); err != nil {
		return err
	}
	return nil
}

// network api
type NetworkApi struct{ ResourceApi }

func (c NeutronV2) Network() NetworkApi {
	return NetworkApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint,
			ResourceUrl: "networks"},
	}
}
func (c NeutronV2) ListRouterPorts(routerId string) (neutron.Ports, error) {
	query := url.Values{}
	query.Set("device_id", routerId)
	return c.Port().List(query)
}
func (c NetworkApi) List(query url.Values) ([]neutron.Network, error) {
	body := struct{ Networks []neutron.Network }{}
	if _, err := c.Get(NETWORKS, query, &body); err != nil {
		return nil, err
	}
	return body.Networks, nil
}
func (c NetworkApi) ListByName(name string) ([]neutron.Network, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c NetworkApi) Show(id string) (*neutron.Network, error) {
	body := struct{ Network neutron.Network }{}
	if _, err := c.Get("networks/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.Network, nil
}
func (c NetworkApi) Found(idOrName string) (*neutron.Network, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c NetworkApi) Create(params map[string]interface{}) (*neutron.Network, error) {
	body := struct{ Network neutron.Network }{}
	if _, err := c.Post(NETWORKS, map[string]interface{}{"network": params}, &body); err != nil {
		return nil, err
	}
	return &body.Network, nil
}

func (c NetworkApi) Delete(id string) error {
	_, err := c.ResourceDelete(id)
	return err
}

// subnet api
type SubnetApi struct{ ResourceApi }

func (c NeutronV2) Subnet() SubnetApi {
	return SubnetApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint, ResourceUrl: "subnets"},
	}
}
func (c SubnetApi) List(query url.Values) ([]neutron.Subnet, error) {
	body := struct{ Subnets []neutron.Subnet }{}
	if _, err := c.Get(SUBNETS, query, &body); err != nil {
		return nil, err
	}
	return body.Subnets, nil
}
func (c SubnetApi) ListByName(name string) ([]neutron.Subnet, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}

func (c SubnetApi) Show(id string) (*neutron.Subnet, error) {
	body := struct{ Subnet neutron.Subnet }{}
	if _, err := c.Get("subnets/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.Subnet, nil
}
func (c SubnetApi) Found(idOrName string) (*neutron.Subnet, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c SubnetApi) Create(params map[string]interface{}) (*neutron.Subnet, error) {
	body := struct{ Subnet neutron.Subnet }{}
	if _, err := c.Post(SUBNETS, map[string]interface{}{"subnet": params}, &body); err != nil {
		return nil, err
	}
	return &body.Subnet, nil
}

func (c SubnetApi) Delete(id string) error {
	_, err := c.ResourceDelete(id)
	return err
}

// port api
type PortApi struct{ ResourceApi }

func (c NeutronV2) Port() PortApi {
	return PortApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint,
			ResourceUrl: "ports"},
	}
}

func (c PortApi) List(query url.Values) ([]neutron.Port, error) {
	body := struct{ Ports []neutron.Port }{}
	if _, err := c.Get(PORTS, query, &body); err != nil {
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
	body := struct{ Port neutron.Port }{}
	if _, err := c.Get("ports/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.Port, nil
}
func (c PortApi) Update(id string, options map[string]interface{}) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.Put("ports/"+id, map[string]interface{}{"port": options}, body); err != nil {
		return nil, err
	}
	return &body.Port, nil
}
func (c PortApi) Found(idOrName string) (*neutron.Port, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c PortApi) Create(params map[string]interface{}) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.Post(PORTS, map[string]interface{}{"port": params}, &body); err != nil {
		return nil, err
	}
	return &body.Port, nil
}

func (c PortApi) Delete(id string) error {
	_, err := c.ResourceDelete(id)
	return err
}

// neutron agent api
type agentApi struct {
	ResourceApi
}

func (c NeutronV2) Agent() agentApi {
	return agentApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
	}
}

func (c agentApi) List(query url.Values) ([]neutron.Agent, error) {
	body := struct{ Agents []neutron.Agent }{}
	if _, err := c.Get("agents", query, &body); err != nil {
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
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
	}
}

func (c sgApi) List(query url.Values) ([]neutron.SecurityGroup, error) {
	body := struct {
		SecurityGroups []neutron.SecurityGroup `json:"security_groups"`
	}{}
	if _, err := c.Get("security-groups", query, &body); err != nil {
		return nil, err
	}
	return body.SecurityGroups, nil
}
func (c sgApi) Show(id string) (*neutron.SecurityGroup, error) {
	body := struct {
		SecurityGroup neutron.SecurityGroup `json:"security_group"`
	}{}
	if _, err := c.Get("security-groups/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.SecurityGroup, nil
}
func (c sgApi) Found(idOrName string) (*neutron.SecurityGroup, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// security group rule api
type sgRuleApi struct {
	ResourceApi
}

func (c NeutronV2) SecurityGroupRule() sgRuleApi {
	return sgRuleApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
	}
}

func (c sgRuleApi) List(query url.Values) ([]neutron.SecurityGroupRule, error) {
	body := struct {
		SecurityGroupRules []neutron.SecurityGroupRule `json:"security_group_rules"`
	}{}
	if _, err := c.Get("security-group-rules", query, &body); err != nil {
		return nil, err
	}
	return body.SecurityGroupRules, nil
}

func (c sgRuleApi) Show(id string) (*neutron.SecurityGroupRule, error) {
	body := struct {
		SecurityGroupRules neutron.SecurityGroupRule `json:"security_group_rule"`
	}{}
	if _, err := c.Get("security-group-rules/"+id, nil, &body); err != nil {
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
		ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
		// ResourceApi: ResourceApi{
		// 	Endpoint:    c.BaseUrl,
		// 	BaseUrl:     "qos/policies",
		// 	Client:      c.session,
		// 	SingularKey: "policies",
		// 	PluralKey:   "policies",
		// },
	}
}
func (c qosPolicyApi) List(query url.Values) ([]neutron.QosPolicy, error) {
	body := struct{ Policies []neutron.QosPolicy }{}
	if _, err := c.Get("qos/policies", query, &body); err != nil {
		return nil, err
	}
	return body.Policies, nil
}
func (c qosPolicyApi) Show(id string) (*neutron.QosPolicy, error) {
	body := struct{ Policy neutron.QosPolicy }{}
	if _, err := c.Get("qos/policies/"+id, nil, &body); err != nil {
		return nil, err
	}
	return &body.Policy, nil
}

func (c qosPolicyApi) Found(idOrName string) (*neutron.QosPolicy, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// qos rule api
type qosRuleApi struct {
	ResourceApi
}

func (c NeutronV2) QosRule() qosRuleApi {
	return qosRuleApi{
		// ResourceApi{Client: c.rawClient, BaseUrl: c.Endpoint},
		// ResourceApi: ResourceApi{
		// 	Endpoint:    c.BaseUrl,
		// 	BaseUrl:     "qos/policies",
		// 	Client:      c.session,
		// 	SingularKey: "policies",
		// 	PluralKey:   "policies",
		// },
	}
}

// func (c qosPolicyApi) List(query url.Values) ([]neutron.QosPolicy, error) {
// 	body := struct{ Policies []neutron.QosPolicy }{}
// 	if _, err := c.SetQuery(query).Get(&body); err != nil {
// 		return nil, err
// 	}
// 	return body.Policies, nil
// }
