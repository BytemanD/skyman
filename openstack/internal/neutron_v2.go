package internal

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
	"github.com/BytemanD/skyman/openstack/result"
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
		result := struct{ Versions model.ApiVersions }{}
		if resp, err := c.Index(nil); err != nil {
			return nil, err
		} else if err := resp.UnmarshalBody(&result); err != nil {
			return nil, err
		}
		versions := result.Versions
		c.currentVersion = versions.Current()
	}
	if c.currentVersion != nil {
		return c.currentVersion, nil
	}
	return nil, fmt.Errorf("current version not found")
}

type routerApi struct{ ResourceApi }
type NetworkApi struct{ ResourceApi }
type SubnetApi struct{ ResourceApi }
type PortApi struct{ ResourceApi }
type agentApi struct{ ResourceApi }
type sgApi struct{ ResourceApi }
type sgRuleApi struct{ ResourceApi }
type qosPolicyApi struct{ ResourceApi }
type qosRuleApi struct{ ResourceApi }

func (c NeutronV2) Router() routerApi {
	return routerApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "routers",
			SingularKey: "router",
			PluralKey:   "routers",
		},
	}
}
func (c NeutronV2) Network() NetworkApi {
	return NetworkApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "networks",
			SingularKey: "network",
			PluralKey:   "networks",
		},
	}
}
func (c NeutronV2) Subnet() SubnetApi {
	return SubnetApi{
		ResourceApi{
			Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "subnets",
			SingularKey: "subnet",
			PluralKey:   "subnets",
		},
	}
}
func (c NeutronV2) Port() PortApi {
	return PortApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "ports",
			SingularKey: "port",
			PluralKey:   "ports",
		},
	}
}
func (c NeutronV2) Agent() agentApi {
	return agentApi{
		ResourceApi{Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "agents",
			SingularKey: "agent",
			PluralKey:   "agents",
		},
	}
}
func (c NeutronV2) SecurityGroupRule() sgRuleApi {
	return sgRuleApi{
		ResourceApi{
			Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "security-group-rules",
			SingularKey: "security_group_rule",
			PluralKey:   "security_group_rules",
		},
	}
}
func (c NeutronV2) SecurityGroup() sgApi {
	return sgApi{
		ResourceApi{
			Client: c.rawClient, BaseUrl: c.Url,
			ResourceUrl: "security-groups",
			SingularKey: "security_group",
			PluralKey:   "security_groups"},
	}
}
func (c NeutronV2) QosPolicy() qosPolicyApi {
	return qosPolicyApi{
		ResourceApi{
			Client:      c.rawClient,
			BaseUrl:     c.Url,
			ResourceUrl: "qos/policies",
			SingularKey: "policy",
			PluralKey:   "policies",
		},
	}
}
func (c NeutronV2) QosRule() qosRuleApi {
	return qosRuleApi{
		ResourceApi: ResourceApi{
			Client:      c.rawClient,
			BaseUrl:     c.Url,
			ResourceUrl: "qos/policies",
			SingularKey: "policies",
			PluralKey:   "policies",
		},
	}
}

// router api

func (c routerApi) List(query url.Values) ([]neutron.Router, error) {
	return ListResource[neutron.Router](c.ResourceApi, query)
}
func (c routerApi) ListByName(name string) ([]neutron.Router, error) {
	return c.List(url.Values{"name": []string{name}})
}

func (c routerApi) Show(id string) (*neutron.Router, error) {
	return ShowResource[neutron.Router](c.ResourceApi, id)
}
func (c routerApi) Create(params map[string]interface{}) (*neutron.Router, error) {
	result := struct{ Router neutron.Router }{}
	if _, err := c.R().SetBody(ReqBody{"router": params}).SetResult(result).Post(); err != nil {
		return nil, err
	}
	return &result.Router, nil
}

func (c routerApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c routerApi) Find(idOrName string) (*neutron.Router, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// Interface: subnet-id | port=port-id
func (c routerApi) AddSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.R().SetBody(body).Put(routerId, "add_router_interface"); err != nil {
		return err
	}
	return nil
}
func (c routerApi) AddPort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.R().SetBody(body).Put(routerId, "add_router_interface"); err != nil {
		return err
	}
	return nil
}
func (c routerApi) RemoveSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.R().SetBody(body).Put(routerId, "remove_router_interface"); err != nil {
		return err
	}
	return nil
}
func (c routerApi) RemovePort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.R().SetBody(body).Put(routerId, "remove_router_interface"); err != nil {
		return err
	}
	return nil
}

// network api

func (c NetworkApi) List(query url.Values) ([]neutron.Network, error) {
	body := struct{ Networks []neutron.Network }{}
	if _, err := c.Get(NETWORKS, query, &body); err != nil {
		return nil, err
	}
	return body.Networks, nil
}
func (c NetworkApi) ListByName(name string) ([]neutron.Network, error) {
	return c.List(url.Values{"name": []string{name}})
}

func (c NetworkApi) Show(id string) (*neutron.Network, error) {
	return ShowResource[neutron.Network](c.ResourceApi, id)
}
func (c NetworkApi) Find(idOrName string) (*neutron.Network, error) {
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

func (c SubnetApi) List(query url.Values) ([]neutron.Subnet, error) {
	return ListResource[neutron.Subnet](c.ResourceApi, query)
}
func (c SubnetApi) ListByName(name string) ([]neutron.Subnet, error) {
	return c.List(url.Values{"name": []string{name}})
}

func (c SubnetApi) Show(id string) (*neutron.Subnet, error) {
	return ShowResource[neutron.Subnet](c.ResourceApi, id)
}
func (c SubnetApi) Find(idOrName string) (*neutron.Subnet, error) {
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

func (c PortApi) List(query url.Values) ([]neutron.Port, error) {
	return ListResource[neutron.Port](c.ResourceApi, query)
}
func (c PortApi) ListByName(name string) ([]neutron.Port, error) {
	return c.List(url.Values{"name": []string{name}})
}
func (c PortApi) ListByDeviceId(deviceId string) (neutron.Ports, error) {
	return c.List(url.Values{"device_id": []string{deviceId}})
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
func (c PortApi) Find(idOrName string) (*neutron.Port, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// neutron agent api

func (c agentApi) List(query url.Values) ([]neutron.Agent, error) {
	return ListResource[neutron.Agent](c.ResourceApi, query)
}

// security group api

func (c sgApi) List(query url.Values) ([]neutron.SecurityGroup, error) {
	return ListResource[neutron.SecurityGroup](c.ResourceApi, query)
}
func (c sgApi) List2(query url.Values) result.ItemsResult[neutron.SecurityGroup] {
	r := result.NewItemsResult[neutron.SecurityGroup](
		c.R().SetQuery(query).Get(),
	)
	fmt.Printf("222222222 %v \n", r)
	r.SetKey(c.ResourceApi.PluralKey)
	return *r
}
func (c sgApi) Show(id string) (*neutron.SecurityGroup, error) {
	return ShowResource[neutron.SecurityGroup](c.ResourceApi, id)
}
func (c sgApi) Find(idOrName string) (*neutron.SecurityGroup, error) {
	return FindResource(idOrName, c.Show, c.List)
}

// security group rule api

func (c sgRuleApi) List(query url.Values) ([]neutron.SecurityGroupRule, error) {
	return ListResource[neutron.SecurityGroupRule](c.ResourceApi, query)
}
func (c sgRuleApi) Show(id string) (*neutron.SecurityGroupRule, error) {
	return ShowResource[neutron.SecurityGroupRule](c.ResourceApi, id)
}

// qos policy api

func (c qosPolicyApi) List(query url.Values) ([]neutron.QosPolicy, error) {
	return ListResource[neutron.QosPolicy](c.ResourceApi, query)
}
func (c qosPolicyApi) Show(id string) (*neutron.QosPolicy, error) {
	return ShowResource[neutron.QosPolicy](c.ResourceApi, id)
}
func (c qosPolicyApi) Find(idOrName string) (*neutron.QosPolicy, error) {
	return FindResource(idOrName, c.Show, c.List)
}

func (c qosRuleApi) List(query url.Values) ([]neutron.QosRule, error) {
	return ListResource[neutron.QosRule](c.ResourceApi, query)
}
