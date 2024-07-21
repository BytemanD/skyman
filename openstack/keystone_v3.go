package openstack

import (
	"net/url"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model/keystone"
	"github.com/BytemanD/skyman/utility"
)

const V3 = "v3"

type roleAssignmentApi struct {
	ResourceApi
}

func (o *Openstack) KeystoneV3() *KeystoneV3 {
	if o.keystoneClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint("identity", "keystone", "public")
		if err != nil {
			logging.Fatal("get keystone endpoint falied: %v", err)
		}
		o.keystoneClient = &KeystoneV3{
			RestClient2: NewRestClient2(utility.VersionUrl(endpoint, V3), o.AuthPlugin),
		}
	}
	return o.keystoneClient
}

func (c KeystoneV3) Endpoint() *endpointApi {
	return &endpointApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "endpoints",
			Client:      c.session,
			SingularKey: "endpoint",
			PluralKey:   "endpoints",
		},
	}
}
func (c KeystoneV3) Service() serviceApi {
	return serviceApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "services",
			Client:      c.session,
			SingularKey: "service",
			PluralKey:   "services",
		},
	}
}
func (c KeystoneV3) User() userApi {
	return userApi{
		ResourceApi: ResourceApi{
			Endpoint: c.BaseUrl,
			BaseUrl:  "users",
			Client:   c.session,
		},
	}
}
func (c KeystoneV3) Project() projectApi {
	return projectApi{
		ResourceApi: ResourceApi{
			Endpoint:    c.BaseUrl,
			BaseUrl:     "projects",
			Client:      c.session,
			SingularKey: "project",
			PluralKey:   "projects",
		},
	}
}
func (c KeystoneV3) RoleAssignment() roleAssignmentApi {
	return roleAssignmentApi{
		ResourceApi: ResourceApi{
			Endpoint: c.BaseUrl,
			BaseUrl:  "role_assignments",
			Client:   c.session,
		},
	}
}
func (c KeystoneV3) Region() regionApi {
	return regionApi{
		ResourceApi: ResourceApi{
			Endpoint: c.BaseUrl,
			BaseUrl:  "regions",
			Client:   c.session,
		},
	}
}

// region api
type regionApi struct {
	ResourceApi
}

func (c regionApi) List(query url.Values) ([]keystone.Region, error) {
	respBody := struct{ Regions []keystone.Region }{}
	if _, err := c.SetQuery(query).Get(&respBody); err != nil {
		return nil, err
	}
	return respBody.Regions, nil
}

// service pi
type serviceApi struct {
	ResourceApi
}

func (c serviceApi) Show(id string) (*keystone.Service, error) {
	body := struct{ Service keystone.Service }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Service, nil
}
func (c serviceApi) List(query url.Values) ([]keystone.Service, error) {
	respBody := struct{ Services []keystone.Service }{}
	if _, err := c.SetQuery(query).Get(&respBody); err != nil {
		return nil, err
	}
	return respBody.Services, nil
}
func (c serviceApi) ListByName(name string) ([]keystone.Service, error) {
	return c.List(utility.UrlValues(map[string]string{"name": name}))
}
func (c serviceApi) ShowByType(t string) (*keystone.Service, error) {
	query := url.Values{}
	query.Set("type", t)
	body := struct{ Services []keystone.Service }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	switch len(body.Services) {
	case 0:
		return nil, ItemNotFoundError("no service with type %s", t)
	case 1:
		return &body.Services[0], nil
	default:
		return nil, MultiItemsError("multi services with type %s", t)
	}
}
func (c serviceApi) ShowByName(n string) (*keystone.Service, error) {
	query := url.Values{}
	query.Set("name", n)
	body := struct{ Services []keystone.Service }{}
	if _, err := c.SetQuery(query).Get(&body); err != nil {
		return nil, err
	}
	switch len(body.Services) {
	case 0:
		return nil, ItemNotFoundError("no service with name %s", n)
	case 1:
		return &body.Services[0], nil
	default:
		return nil, MultiItemsError("multi services with name %s", n)
	}
}
func (c serviceApi) Found(idOrName string) (*keystone.Service, error) {
	return FoundResource[keystone.Service](c.ResourceApi, idOrName)
}
func (c serviceApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

// endpoint api
type endpointApi struct {
	ResourceApi
}

func (c endpointApi) List(query url.Values) ([]keystone.Endpoint, error) {
	respBody := struct{ Endpoints []keystone.Endpoint }{}
	if _, err := c.SetQuery(query).Get(&respBody); err != nil {
		return nil, err
	}
	return respBody.Endpoints, nil
}

func (c endpointApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}
func (c endpointApi) ListByService(service_id string) ([]keystone.Endpoint, error) {
	return c.List(utility.UrlValues(map[string]string{"service_id": service_id}))
}

// project api
type projectApi struct {
	ResourceApi
}

func (c projectApi) List(query url.Values) ([]auth.Project, error) {
	body := struct{ Projects []auth.Project }{}
	if _, err := c.Get(&body); err != nil {
		return body.Projects, nil
	}
	return body.Projects, nil
}
func (c projectApi) Show(id string) (*auth.Project, error) {
	body := struct{ Project auth.Project }{}
	if _, err := c.AppendUrl(id).Get(&body); err != nil {
		return nil, err
	}
	return &body.Project, nil
}

func (c projectApi) Delete(id string) error {
	_, err := c.AppendUrl(id).Delete(nil)
	return err
}

func (c projectApi) Found(idOrName string) (*auth.Project, error) {
	return FoundResource[auth.Project](c.ResourceApi, idOrName)
}

// user api
type userApi struct {
	ResourceApi
}
