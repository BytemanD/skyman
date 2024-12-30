package internal

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

type RegionApi struct{ ResourceApi }
type ServiceApi struct{ ResourceApi }
type EndpointApi struct{ ResourceApi }
type ProjectApi struct{ ResourceApi }
type UserApi struct{ ResourceApi }
type RoleApi struct{ ResourceApi }

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
func (c ServiceApi) Find(idOrName string) (*keystone.Service, error) {
	return FindResource(idOrName, c.Show, c.List)
}
func (c ServiceApi) Delete(id string) error {
	_, err := DeleteResource(c.ResourceApi, id)
	return err
}
