package identity

import (
	"net/url"
)

func (client IdentityClientV3) ServiceList(query url.Values) ([]Service, error) {
	body := map[string][]Service{"services": {}}
	err := client.List("services", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["services"], nil
}
func (client IdentityClientV3) ServiceListByName(name string) ([]Service, error) {
	query := url.Values{}
	query.Add("name", name)
	return client.ServiceList(query)
}
func (client IdentityClientV3) ServiceShow(serviceId string) (*Service, error) {
	body := map[string]*Service{"service": {}}
	err := client.Show("services", serviceId, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["service"], nil
}
func (client IdentityClientV3) EndpointList(query url.Values) ([]Endpoint, error) {
	body := map[string][]Endpoint{"endpoints": {}}
	err := client.List("endpoints", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["endpoints"], nil
}
func (client IdentityClientV3) UserList(query url.Values) ([]User, error) {
	body := map[string][]User{"users": {}}
	err := client.List("users", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["users"], nil
}
func (client IdentityClientV3) UserShow(userId string) (*User, error) {
	body := map[string]*User{"user": {}}
	err := client.Show("users", userId, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["user"], nil
}
func (client IdentityClientV3) ProjectList(query url.Values) ([]Project, error) {
	body := map[string][]Project{"projects": {}}
	err := client.List("projects", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["projects"], nil
}
func (client IdentityClientV3) RoleAssignmentList(query url.Values) ([]RoleAssigment, error) {
	body := map[string][]RoleAssigment{"role_assignments": {}}
	err := client.List("role_assignments", query, nil, &body)
	if err != nil {
		return nil, err
	}
	return body["role_assignments"], nil
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
