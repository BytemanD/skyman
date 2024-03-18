package openstack

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

const V3 = "v3"

type KeystoneV3 struct {
	RestClient
}
type EndpointApi struct {
	KeystoneV3
}
type ServiceApi struct {
	KeystoneV3
}
type UserApi struct {
	KeystoneV3
}
type ProjectApi struct {
	KeystoneV3
}
type RoleAssignmentApi struct {
	KeystoneV3
}
type RegionApi struct {
	KeystoneV3
}

func (o *Openstack) KeystoneV3() *KeystoneV3 {
	if o.keystoneClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint("identity", "keystone", "public")
		if err != nil {
			logging.Fatal("get keystone endpoint falied: %v", err)
		}
		o.keystoneClient = &KeystoneV3{
			RestClient{
				BaseUrl:    utility.VersionUrl(endpoint, V3),
				AuthPlugin: o.AuthPlugin},
		}
	}
	return o.keystoneClient
}

func (c KeystoneV3) Endpoints() EndpointApi {
	return EndpointApi{c}
}
func (c KeystoneV3) Services() ServiceApi {
	return ServiceApi{c}
}
func (c KeystoneV3) Users() UserApi {
	return UserApi{c}
}
func (c KeystoneV3) Projects() ProjectApi {
	return ProjectApi{c}
}
func (c KeystoneV3) RoleAssignments() RoleAssignmentApi {
	return RoleAssignmentApi{c}
}
func (c KeystoneV3) Regions() RegionApi {
	return RegionApi{c}
}

func (c KeystoneV3) GetCurrentVersion() (*model.ApiVersion, error) {
	resp, err := c.RestClient.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]model.ApiVersions{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Current(), nil
}
func (c KeystoneV3) GetStableVersion() (*model.ApiVersion, error) {
	resp, err := c.RestClient.Index()
	if err != nil {
		return nil, err
	}
	type apiVersion struct {
		Values model.ApiVersions `json:"values"`
	}
	versions := map[string]apiVersion{"versions": {}}
	resp.BodyUnmarshal(&versions)
	return versions["versions"].Values.Stable(), nil
}

func (c EndpointApi) List(query url.Values) ([]keystone.Endpoint, error) {
	resp, err := c.KeystoneV3.Get("endpoints", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Endpoints []keystone.Endpoint }{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Endpoints, nil
}
func (c EndpointApi) ListByService(serviceType string, serviceName string, serviceInterface string) ([]keystone.Endpoint, error) {
	services, err := c.Services().List(utility.UrlValues(map[string]string{
		"type":      serviceType,
		"name":      serviceName,
		"interface": serviceInterface,
	}))
	if err != nil {
		return nil, err
	}
	if len(services) == 0 {
		return nil, fmt.Errorf("service %s:%s:%s not found", serviceType, serviceName, serviceInterface)
	}
	endpoints, err := c.List(utility.UrlValues(map[string]string{
		"service_id": services[0].Resource.Id,
	}))
	if err != nil {
		return nil, err
	}
	return endpoints, err

}

func (c ServiceApi) List(query url.Values) ([]keystone.Service, error) {
	resp, err := c.KeystoneV3.Get("services", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Services []keystone.Service }{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Services, nil
}
func (c ServiceApi) ListByName(name string) ([]keystone.Service, error) {
	return c.List(utility.UrlValues(map[string]string{
		"name": name,
	}))
}
func (c ServiceApi) Show(id string) (*keystone.Service, error) {
	resp, err := c.KeystoneV3.Get(utility.UrlJoin("services", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*keystone.Service{"service": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["service"], nil
}
func (c UserApi) List(query url.Values) ([]auth.User, error) {
	resp, err := c.KeystoneV3.Get("users", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Users []auth.User }{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Users, nil
}

func (c UserApi) ListByProjectId(projectId string) ([]auth.User, error) {
	items, err := c.RoleAssignments().List(utility.UrlValues(map[string]string{
		"scope.project.id": projectId,
	}))
	if err != nil {
		return nil, err
	}
	users := []auth.User{}
	for _, roleAssignment := range items {
		user, err := c.Show(roleAssignment.User.Id)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, nil
}

func (c UserApi) Show(id string) (*auth.User, error) {
	resp, err := c.KeystoneV3.Get(utility.UrlJoin("users", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*auth.User{"users": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["users"], nil
}

func (c ProjectApi) List(query url.Values) ([]auth.Project, error) {
	resp, err := c.KeystoneV3.Get("projects", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Projects []auth.Project }{}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body.Projects, nil
}

func (c ProjectApi) Show(id string) (*auth.Project, error) {
	resp, err := c.KeystoneV3.Get(utility.UrlJoin("projects", id), nil)
	if err != nil {
		return nil, err
	}
	body := map[string]*auth.Project{"projects": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["projects"], nil
}
func (c ProjectApi) Delete(id string) error {
	_, err := c.KeystoneV3.Delete(utility.UrlJoin("projects", id), nil)
	return err
}
func (c RoleAssignmentApi) List(query url.Values) ([]keystone.RoleAssigment, error) {
	resp, err := c.KeystoneV3.Get("role_assignments", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]keystone.RoleAssigment{"role_assignments": {}}
	if err := resp.BodyUnmarshal(&body); err != nil {
		return nil, err
	}
	return body["role_assignments"], nil
}
func (c RegionApi) List(query url.Values) ([]keystone.Region, error) {
	resp, err := c.KeystoneV3.Get("regions", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Regions []keystone.Region }{}
	err = resp.BodyUnmarshal(&body)
	return body.Regions, err
}
