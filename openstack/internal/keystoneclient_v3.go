package internal

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

type KeystoneV3 struct{ *ServiceClient }

func (c KeystoneV3) GetStableVersion() (*model.ApiVersion, error) {
	result := struct {
		Versions map[string]model.ApiVersions `json:"versions"`
	}{Versions: map[string]model.ApiVersions{}}
	if _, err := c.Index(&result); err != nil {
		return nil, err
	}
	if len(result.Versions["values"]) > 0 {
		v := result.Versions["values"][0]
		return &v, nil
	} else {
		return nil, fmt.Errorf("stable version not found")
	}
}

func (c KeystoneV3) ListRegion(query url.Values) ([]keystone.Region, error) {
	return QueryResource[keystone.Region](c.ServiceClient, URL_REGIONS.F(), query, "regions")
}

func (c KeystoneV3) ListService(query url.Values) ([]keystone.Service, error) {
	return QueryResource[keystone.Service](c.ServiceClient, URL_SERVICES.F(), query, SERVICES)
}
func (c KeystoneV3) GetService(id string) (*keystone.Service, error) {
	return GetResource[keystone.Service](c.ServiceClient, URL_SERVER.F(id), "service")
}
func (c KeystoneV3) ListByName(name string) ([]keystone.Service, error) {
	return c.ListService(utility.UrlValues(map[string]string{"name": name}))
}
func (c KeystoneV3) GetServiceByType(t string) (*keystone.Service, error) {
	services, err := c.ListService(url.Values{"type": []string{t}})
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
func (c KeystoneV3) GetServiceByName(t string) (*keystone.Service, error) {
	services, err := c.ListService(url.Values{"name": []string{t}})
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
func (c KeystoneV3) CreateService(service keystone.Service) (*keystone.Service, error) {
	type serviceBody struct {
		Service keystone.Service `json:"service"`
	}
	reqBody, _ := json.Marshal(serviceBody{Service: service})
	respBody := serviceBody{Service: keystone.Service{}}
	_, err := c.R().SetBody(reqBody).SetResult(&respBody).Post(URL_SERVICES.F())
	return &respBody.Service, err
}
func (c KeystoneV3) FindService(idOrName string) (*keystone.Service, error) {
	return QueryByIdOrName(idOrName, c.GetService, c.ListService)
}
func (c KeystoneV3) DeleteService(id string) error {
	return DeleteResource(c.ServiceClient, URL_SERVICE.F(id))
}
func (c KeystoneV3) ListEndpoint(query url.Values) ([]keystone.Endpoint, error) {
	return QueryResource[keystone.Endpoint](c.ServiceClient, URL_ENDPOINTS.F(), query, "endpoints")
}

func (c KeystoneV3) DeleteEndpoint(id string) error {
	return DeleteResource(c.ServiceClient, URL_ENDPOINT.F(id))
}
func (c KeystoneV3) ListEndpointByService(service_id string) ([]keystone.Endpoint, error) {
	return c.ListEndpoint(utility.UrlValues(map[string]string{"service_id": service_id}))
}
func (c KeystoneV3) CreateEndpoint(endpoint keystone.Endpoint) (*keystone.Endpoint, error) {
	type Body struct {
		Endpoint keystone.Endpoint `json:"endpoint"`
	}
	reqBody, _ := json.Marshal(Body{Endpoint: endpoint})
	respBody := Body{Endpoint: keystone.Endpoint{}}
	_, err := c.R().SetBody(reqBody).SetResult(&respBody).Post(URL_ENDPOINTS.F())
	return &respBody.Endpoint, err
}
func (c KeystoneV3) ListProject(query url.Values) ([]model.Project, error) {
	return QueryResource[model.Project](c.ServiceClient, URL_PROJECTS.F(), query, "projects")
}
func (c KeystoneV3) GetProject(id string) (*model.Project, error) {
	return GetResource[model.Project](c.ServiceClient, URL_PROJECT.F(id), "project")
}

func (c KeystoneV3) DeleteProject(id string) error {
	return DeleteResource(c.ServiceClient, URL_PROJECT.F(id))
}

func (c KeystoneV3) FindProject(idOrName string) (*model.Project, error) {
	return QueryByIdOrName(idOrName, c.GetProject, c.ListProject)
}

// user api
func (c KeystoneV3) ListUser(query url.Values) ([]model.User, error) {
	return QueryResource[model.User](c.ServiceClient, URL_USERS.F(), query, USERS)
}

func (c KeystoneV3) GetUser(id string) (*model.User, error) {
	return GetResource[model.User](c.ServiceClient, URL_USER.F(id), USER)

}
func (c KeystoneV3) FindUser(idOrName string) (*model.User, error) {
	return QueryByIdOrName(idOrName, c.GetUser, c.ListUser)
}
func (c KeystoneV3) ListRoleAssigment(query url.Values) ([]keystone.RoleAssigment, error) {
	return QueryResource[keystone.RoleAssigment](
		c.ServiceClient, URL_ROLE_ASSIGNMENTS.F(), query, "role_assignments")
}

func (c KeystoneV3) ListUsersByProject(projectId string) ([]model.User, error) {
	items, err := c.ListRoleAssigment(
		utility.UrlValues(map[string]string{"scope.project.id": projectId}))
	if err != nil {
		return nil, err
	}
	users := []model.User{}
	for _, roleAssignment := range items {
		user, err := c.GetUser(roleAssignment.User.Id)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, nil
}
