package internal

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

type RegionApi struct{ ResourceApi }
type ServiceApi struct{ ResourceApi }
type EndpointApi struct{ ResourceApi }
type ProjectApi struct{ ResourceApi }
type UserApi struct{ ResourceApi }
type RoleApi struct{ ResourceApi }
type RoleAssignmentApi struct{ ResourceApi }

func (c RegionApi) List(query url.Values) ([]keystone.Region, error) {
	respBody := struct{ Regions []keystone.Region }{}
	if _, err := c.Get("regions", query, &respBody); err != nil {
		return nil, err
	}
	return respBody.Regions, nil
}

func (c ServiceApi) List(query url.Values) ([]keystone.Service, error) {
	return ListResource[keystone.Service](c.ResourceApi, query)
}
func (c ServiceApi) Show(id string) (*keystone.Service, error) {
	return ShowResource[keystone.Service](c.ResourceApi, id)
}
func (c ServiceApi) ListByName(name string) ([]keystone.Service, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}
func (c ServiceApi) ShowByType(t string) (*keystone.Service, error) {
	services, err := c.List(url.Values{"type": []string{t}})
	if err != nil {
		return nil, err
	}
	switch len(services) {
	case 0:
		//TODO
		return nil, fmt.Errorf("no service with type %s", t)
	case 1:
		return &services[0], nil
	default:
		return nil, fmt.Errorf("multi services with type %s", t)
	}
}
func (c ServiceApi) ShowByName(t string) (*keystone.Service, error) {
	services, err := c.List(url.Values{"name": []string{t}})
	if err != nil {
		return nil, err
	}
	switch len(services) {
	case 0:
		//TODO
		return nil, fmt.Errorf("no service with name %s", t)
	case 1:
		return &services[0], nil
	default:
		return nil, fmt.Errorf("multi services with name %s", t)
	}
}
func (c ServiceApi) Create(service keystone.Service) (*keystone.Service, error) {
	type serviceBody struct {
		Service keystone.Service `json:"service"`
	}

	reqBody, _ := json.Marshal(serviceBody{Service: service})
	resp, err := c.Post("services", reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := serviceBody{Service: keystone.Service{}}
	json.Unmarshal(resp.Body(), &respBody)
	return &respBody.Service, nil
}
func (c ServiceApi) Find(idOrName string) (*keystone.Service, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c ServiceApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c EndpointApi) List(query url.Values) ([]keystone.Endpoint, error) {
	return ListResource[keystone.Endpoint](c.ResourceApi, query)
}

func (c EndpointApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
func (c EndpointApi) ListByService(service_id string) ([]keystone.Endpoint, error) {
	return c.List(utility.UrlValues(map[string]string{"service_id": service_id}))
}
func (c EndpointApi) Create(endpoint keystone.Endpoint) (*keystone.Endpoint, error) {
	type Body struct {
		Endpoint keystone.Endpoint `json:"endpoint"`
	}

	reqBody, _ := json.Marshal(Body{Endpoint: endpoint})
	resp, err := c.Post("endpoints", reqBody, nil)
	if err != nil {
		return nil, err
	}
	respBody := Body{Endpoint: keystone.Endpoint{}}
	json.Unmarshal(resp.Body(), &respBody)
	return &respBody.Endpoint, nil
}
func (c ProjectApi) List(query url.Values) ([]model.Project, error) {
	return ListResource[model.Project](c.ResourceApi, query)
}
func (c ProjectApi) Show(id string) (*model.Project, error) {
	return ShowResource[model.Project](c.ResourceApi, id)
}

func (c ProjectApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}

func (c ProjectApi) Find(idOrName string) (*model.Project, error) {
	return FindResource[model.Project](idOrName, c.Show, c.List)
}

// user api
func (c UserApi) List(query url.Values) ([]model.User, error) {
	return ListResource[model.User](c.ResourceApi, query)
}

func (c UserApi) Show(id string) (*model.User, error) {
	return ShowResource[model.User](c.ResourceApi, id)
}
func (c UserApi) Find(idOrName string) (*model.User, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c RoleAssignmentApi) List(query url.Values) ([]keystone.RoleAssigment, error) {
	return ListResource[keystone.RoleAssigment](c.ResourceApi, query)
}

type KeystoneV3 struct {
	*ServiceClient
}

func (c KeystoneV3) ListUsersByProjectId(projectId string) ([]model.User, error) {
	items, err := c.RoleAssignment().List(
		utility.UrlValues(map[string]string{"scope.project.id": projectId}))
	if err != nil {
		return nil, err
	}
	users := []model.User{}
	for _, roleAssignment := range items {
		user, err := c.User().Show(roleAssignment.User.Id)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, nil
}

func (c KeystoneV3) GetStableVersion() (*model.ApiVersion, error) {
	respBody := struct {
		Versions map[string]model.ApiVersions `json:"versions"`
	}{}
	if resp, err := c.Index(nil); err != nil {
		return nil, err
	} else {
		if err := resp.UnmarshalBody(&respBody); err != nil {
			return nil, err
		}
	}
	versions := respBody.Versions["values"]
	return versions.Stable(), nil
}
func (c KeystoneV3) Region() *RegionApi {
	return &RegionApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "regions",
			SingularKey: "region",
			PluralKey:   "regions",
		},
	}
}

func (c KeystoneV3) Service() *ServiceApi {
	return &ServiceApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "services",
			SingularKey: "service",
			PluralKey:   "services",
		},
	}
}

func (c KeystoneV3) Endpoint() *EndpointApi {
	return &EndpointApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "endpoints",
			SingularKey: "endpoint",
			PluralKey:   "endpoints",
		},
	}
}

func (c KeystoneV3) User() UserApi {
	return UserApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "users",
			SingularKey: "user",
			PluralKey:   "users",
		},
	}
}
func (c KeystoneV3) Project() ProjectApi {
	return ProjectApi{
		ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "projects",
			SingularKey: "project",
			PluralKey:   "projects",
		},
	}
}
func (c KeystoneV3) RoleAssignment() RoleAssignmentApi {
	return RoleAssignmentApi{
		ResourceApi: ResourceApi{
			Client:      c.Client,
			BaseUrl:     c.BaserUrl(),
			ResourceUrl: "role_assignments",
			SingularKey: "role_assignment",
			PluralKey:   "role_assignments",
		},
	}
}
