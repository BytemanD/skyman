package identity

import (
	"net/url"

	"github.com/BytemanD/skyman/openstack/common"
)

func (client IdentityClientV3) newRequest(resource string, id string, query url.Values, body []byte) common.RestfulRequest {
	return common.RestfulRequest{
		Endpoint: client.endpoint,
		Resource: resource, Id: id,
		Query:   query,
		Body:    body,
		Headers: client.BaseHeaders}
}
func (client IdentityClientV3) ServiceList(query url.Values) ([]Service, error) {
	resp, err := client.Request(client.newRequest("services", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Service{"services": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["services"], nil
}
func (client IdentityClientV3) ServiceListByName(name string) ([]Service, error) {
	query := url.Values{}
	query.Add("name", name)
	return client.ServiceList(query)
}
func (client IdentityClientV3) ServiceShow(serviceId string) (*Service, error) {
	resp, err := client.Request(client.newRequest("services", serviceId, nil, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string]*Service{"service": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["service"], nil
}

func (client IdentityClientV3) EndpointList(query url.Values) ([]Endpoint, error) {
	resp, err := client.Request(client.newRequest("endpoints", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Endpoint{"endpoints": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["endpoints"], nil
}
func (client IdentityClientV3) GetServiceEndpoint(serviceType string, serviceName string, serviceInterface string,
) (string, error) {
	return client.Auth.GetServiceEndpoint(serviceType, serviceName, serviceInterface)
}

func (client IdentityClientV3) UserList(query url.Values) ([]User, error) {
	resp, err := client.Request(client.newRequest("users", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]User{"users": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["users"], nil
}
func (client IdentityClientV3) UserShow(userId string) (*User, error) {
	resp, err := client.Request(client.newRequest("users", userId, nil, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string]*User{"user": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["user"], nil
}

func (client IdentityClientV3) ProjectList(query url.Values) ([]Project, error) {
	resp, err := client.Request(client.newRequest("projects", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Project{"projects": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["projects"], nil
}

func (client IdentityClientV3) ProjectDelete(id string) error {
	_, err := client.Request(
		common.NewResourceDeleteRequest(client.endpoint,
			"projects", id, client.BaseHeaders),
	)
	return err
}

func (client IdentityClientV3) RoleAssignmentList(query url.Values) ([]RoleAssigment, error) {
	resp, err := client.Request(client.newRequest("role_assignments", "", query, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]RoleAssigment{"role_assignments": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["role_assignments"], nil
}
func (client IdentityClientV3) UserListByProjectId(projectId string) ([]User, error) {
	query := url.Values{}
	query.Set("scope.project.id", projectId)
	roleAssignments, err := client.RoleAssignmentList(query)
	if err != nil {
		return nil, err
	}
	users := []User{}
	for _, roleAssignment := range roleAssignments {
		user, err := client.UserShow(roleAssignment.User.Id)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, nil
}

func (client IdentityClientV3) RegionList() ([]Region, error) {
	resp, err := client.Request(client.newRequest("regions", "", nil, nil))
	if err != nil {
		return nil, err
	}
	respBody := map[string][]Region{"regions": {}}
	resp.BodyUnmarshal(&respBody)
	return respBody["regions"], nil
}
