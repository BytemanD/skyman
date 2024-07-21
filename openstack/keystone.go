package openstack

import (
	"encoding/json"
	"net/url"

	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

type KeystoneV3 struct {
	RestClient2
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
	resp, err := c.RestClient2.Index()
	if err != nil {
		return nil, err
	}
	versions := map[string]model.ApiVersions{"versions": {}}
	json.Unmarshal(resp.Body(), &versions)
	return versions["versions"].Current(), nil
}
func (c KeystoneV3) GetStableVersion() (*model.ApiVersion, error) {
	resp, err := c.RestClient2.Index()
	if err != nil {
		return nil, err
	}
	type apiVersion struct {
		Values model.ApiVersions `json:"values"`
	}
	versions := map[string]apiVersion{"versions": {}}
	json.Unmarshal(resp.Body(), &versions)
	return versions["versions"].Values.Stable(), nil
}

func (c EndpointApi) List(query url.Values) ([]keystone.Endpoint, error) {
	resp, err := c.KeystoneV3.Get("endpoints", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Endpoints []keystone.Endpoint }{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	return body.Endpoints, nil
}
func (c EndpointApi) Create(endpoint keystone.Endpoint) (*keystone.Endpoint, error) {
	type Body struct {
		Endpoint keystone.Endpoint `json:"endpoint"`
	}

	reqBody, _ := json.Marshal(Body{Endpoint: endpoint})
	resp, err := c.KeystoneV3.session.Post("endpoints", reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := Body{Endpoint: keystone.Endpoint{}}
	json.Unmarshal(resp.Body(), &respBody)
	return &respBody.Endpoint, nil
}

func (c ServiceApi) Create(service keystone.Service) (*keystone.Service, error) {
	type serviceBody struct {
		Service keystone.Service `json:"service"`
	}

	reqBody, _ := json.Marshal(serviceBody{Service: service})
	resp, err := c.KeystoneV3.session.Post("services", reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := serviceBody{Service: keystone.Service{}}
	json.Unmarshal(resp.Body(), &respBody)
	return &respBody.Service, nil
}

func (c UserApi) List(query url.Values) ([]auth.User, error) {
	resp, err := c.KeystoneV3.Get("users", query)
	if err != nil {
		return nil, err
	}
	body := struct{ Users []auth.User }{}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
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
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	return body["users"], nil
}

func (c RoleAssignmentApi) List(query url.Values) ([]keystone.RoleAssigment, error) {
	resp, err := c.KeystoneV3.Get("role_assignments", query)
	if err != nil {
		return nil, err
	}
	body := map[string][]keystone.RoleAssigment{"role_assignments": {}}
	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return nil, err
	}
	return body["role_assignments"], nil
}
