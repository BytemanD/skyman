package internal

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/neutron"
)

type NeutronV2 struct {
	*ServiceClient
	currentVersion *model.ApiVersion
}

func (c *NeutronV2) GetCurrentVersion() (*model.ApiVersion, error) {
	if c.currentVersion == nil {
		result := struct{ Versions model.ApiVersions }{}
		if _, err := c.Index(&result); err != nil {
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

// router api

func (c NeutronV2) ListRouter(query url.Values) ([]neutron.Router, error) {
	return QueryResource[neutron.Router](c.ServiceClient, URL_ROUTERS.F(), query, ROUTERS)
}
func (c NeutronV2) ListRouterByName(name string) ([]neutron.Router, error) {
	return c.ListRouter(url.Values{"name": []string{name}})
}

func (c NeutronV2) GetRouter(id string) (*neutron.Router, error) {
	return GetResource[neutron.Router](c.ServiceClient, URL_ROUTER.F(id), ROUTER)
}
func (c NeutronV2) CreateRouter(params map[string]any) (*neutron.Router, error) {
	result := struct {
		Router neutron.Router `json:"router"`
	}{}
	if _, err := c.R().SetBody(ReqBody{"router": params}).SetResult(&result).
		Post(URL_ROUTERS.F()); err != nil {
		return nil, err
	}
	return &result.Router, nil
}

func (c NeutronV2) DeleteRouter(id string) error {
	return DeleteResource(c.ServiceClient, URL_ROUTER.F(id))
}
func (c NeutronV2) FindRouter(idOrName string) (*neutron.Router, error) {
	return QueryByIdOrName(idOrName, c.GetRouter, c.ListRouter)
}

func (c NeutronV2) AddRouterSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.R().SetBody(body).Put(URL_ROUTER_ADD_INTERFACE.F(routerId)); err != nil {
		return err
	}
	return nil
}
func (c NeutronV2) AddRouterPort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.R().SetBody(body).Put(URL_ROUTER_ADD_INTERFACE.F(routerId)); err != nil {
		return err
	}
	return nil
}
func (c NeutronV2) RemoveRouterSubnet(routerId, subnetId string) error {
	body := map[string]string{"subnet_id": subnetId}
	if _, err := c.R().SetBody(body).Put(URL_ROUTER_REMOVE_INTERFACE.F(routerId)); err != nil {
		return err
	}
	return nil
}
func (c NeutronV2) RemoveRouterPort(routerId, portId string) error {
	body := map[string]string{"port_id": portId}
	if _, err := c.R().SetBody(body).Put(URL_ROUTER_REMOVE_INTERFACE.F(routerId)); err != nil {
		return err
	}
	return nil
}

// network api

func (c NeutronV2) ListNetwork(query url.Values) ([]neutron.Network, error) {
	return QueryResource[neutron.Network](c.ServiceClient, URL_NETWORKS.F(), query, NETWORKS)
}
func (c NeutronV2) ListNetwrkByName(name string) ([]neutron.Network, error) {
	return c.ListNetwork(url.Values{"name": []string{name}})
}

func (c NeutronV2) GetNetwork(id string) (*neutron.Network, error) {
	return GetResource[neutron.Network](c.ServiceClient, URL_NETWORK.F(id), NETWORK)
}
func (c NeutronV2) FindNetwork(idOrName string) (*neutron.Network, error) {
	return QueryByIdOrName(idOrName, c.GetNetwork, c.ListNetwork)
}
func (c NeutronV2) CreateNetwork(params map[string]any) (*neutron.Network, error) {
	body := struct{ Network neutron.Network }{}
	if _, err := c.R().SetBody(map[string]any{"network": params}).
		SetResult(&body).Post(URL_NETWORKS.F()); err != nil {
		return nil, err
	}
	return &body.Network, nil
}

func (c NeutronV2) DeleteNetwork(id string) error {
	return DeleteResource(c.ServiceClient, URL_NETWORK.F(id))
}

// subnet api

func (c NeutronV2) ListSubnet(query url.Values) ([]neutron.Subnet, error) {
	return QueryResource[neutron.Subnet](c.ServiceClient, URL_SUBNETS.F(), query, SUBNETS)
}
func (c NeutronV2) ListSubnetByName(name string) ([]neutron.Subnet, error) {
	return c.ListSubnet(url.Values{"name": []string{name}})
}

func (c NeutronV2) GetSubnet(id string) (*neutron.Subnet, error) {
	return GetResource[neutron.Subnet](c.ServiceClient, URL_SUBNET.F(id), SUBNET)
}
func (c NeutronV2) FindSubnet(idOrName string) (*neutron.Subnet, error) {
	return QueryByIdOrName(idOrName, c.GetSubnet, c.ListSubnet)
}
func (c NeutronV2) CreateSubnet(params map[string]any) (*neutron.Subnet, error) {
	body := struct{ Subnet neutron.Subnet }{}
	if _, err := c.R().SetBody(map[string]any{"subnet": params}).SetResult(&body).
		Post(URL_SUBNETS.F()); err != nil {
		return nil, err
	}
	return &body.Subnet, nil
}
func (c NeutronV2) DeleteSubnet(id string) error {
	return DeleteResource(c.ServiceClient, URL_SUBNET.F(id))
}

// port api

func (c NeutronV2) ListPort(query url.Values) ([]neutron.Port, error) {
	return QueryResource[neutron.Port](c.ServiceClient, URL_PORTS.F(), query, PORTS)
}
func (c NeutronV2) ListPortByName(name string) ([]neutron.Port, error) {
	return c.ListPort(url.Values{"name": []string{name}})
}
func (c NeutronV2) ListPortByDeviceId(deviceId string) (neutron.Ports, error) {
	return c.ListPort(url.Values{"device_id": []string{deviceId}})
}
func (c NeutronV2) GetPort(id string) (*neutron.Port, error) {
	return GetResource[neutron.Port](c.ServiceClient, URL_PORT.F(id), PORT)
}
func (c NeutronV2) UpdatePort(id string, options map[string]any) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.R().SetBody(map[string]any{"port": options}).
		SetResult(&body).Put(URL_PORT.F(id)); err != nil {
		return nil, err
	}
	return &body.Port, nil
}
func (c NeutronV2) CreatePort(params map[string]any) (*neutron.Port, error) {
	body := struct{ Port neutron.Port }{}
	if _, err := c.R().SetBody(map[string]any{"port": params}).SetResult(&body).
		Post(URL_PORTS.F()); err != nil {
		return nil, err
	}
	return &body.Port, nil
}

func (c NeutronV2) DeletePort(id string) error {
	return DeleteResource(c.ServiceClient, URL_PORT.F(id))
}
func (c NeutronV2) FindPort(idOrName string) (*neutron.Port, error) {
	return QueryByIdOrName(idOrName, c.GetPort, c.ListPort)
}

// neutron agent api

func (c NeutronV2) ListAgent(query url.Values) ([]neutron.Agent, error) {
	return QueryResource[neutron.Agent](c.ServiceClient, URL_AGENTS.F(), query, "agents")
}

// security group api

func (c NeutronV2) ListSecurityGroup(query url.Values) ([]neutron.SecurityGroup, error) {
	return QueryResource[neutron.SecurityGroup](c.ServiceClient, URL_SECURITY_GROUPS.F(), query, "security_groups")
}
func (c NeutronV2) GetSecurityGroup(id string) (*neutron.SecurityGroup, error) {
	return GetResource[neutron.SecurityGroup](c.ServiceClient, URL_SECURITY_GROUP.F(id), "security_group")
}
func (c NeutronV2) FindSecurityGroup(idOrName string) (*neutron.SecurityGroup, error) {
	return QueryByIdOrName(idOrName, c.GetSecurityGroup, c.ListSecurityGroup)
}

// security group rule api

func (c NeutronV2) ListSecurityGroupRule(query url.Values) ([]neutron.SecurityGroupRule, error) {
	return QueryResource[neutron.SecurityGroupRule](
		c.ServiceClient, URL_SECURITY_GROUP_RULES.F(), query, "security_group_rules",
	)
}
func (c NeutronV2) GetSecurityGroupRule(id string) (*neutron.SecurityGroupRule, error) {
	return GetResource[neutron.SecurityGroupRule](
		c.ServiceClient, URL_SECURITY_GROUP_RULES.F(id), "security_group_rule")
}

// qos policy api

func (c NeutronV2) ListQosPolicy(query url.Values) ([]neutron.QosPolicy, error) {
	return QueryResource[neutron.QosPolicy](c.ServiceClient, URL_QOS_POLICIES.F(), query, "policies")
}
func (c NeutronV2) GetQosPolicy(id string) (*neutron.QosPolicy, error) {
	return GetResource[neutron.QosPolicy](c.ServiceClient, URL_QOS_POLICY.F(), "qos_policy")
}
func (c NeutronV2) FindQosPolicy(idOrName string) (*neutron.QosPolicy, error) {
	return QueryByIdOrName(idOrName, c.GetQosPolicy, c.ListQosPolicy)
}

func (c NeutronV2) ListQosRule(query url.Values) ([]neutron.QosRule, error) {
	return QueryResource[neutron.QosRule](c.ServiceClient, URL_QOS_POLICY_RULES.F(), query, "policie_rules")
}
