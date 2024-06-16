/*
OpenStack Client with Golang
*/
package openstack

import (
	"fmt"
	"net/url"

	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
)

type RestClient2 struct {
	BaseUrl string
	session *httpclient.Client
}

func (rest *RestClient2) makeUrl(url string) string {
	return utility.UrlJoin(rest.BaseUrl, url)
}
func (rest *RestClient2) mustHasBaseUrl() error {
	if rest.BaseUrl == "" {
		return fmt.Errorf("base url is required")
	}
	return nil
}
func (rest *RestClient2) AddBaseHeader(key, value string) {
	rest.session.BaseHeaders[key] = value
}
func (rest *RestClient2) Index() (*httpclient.Response, error) {
	parsed, err := url.Parse(rest.BaseUrl)
	if err != nil {
		return nil, err
	}
	return rest.session.Get(
		fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), nil, nil)
}
func (rest *RestClient2) Get(url string, query url.Values) (*httpclient.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Get(rest.makeUrl(url), query, nil)
}
func (rest *RestClient2) Post(url string, body []byte, headers map[string]string) (*httpclient.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Post(rest.makeUrl(url), body, headers)
}
func (rest *RestClient2) Put(url string, body []byte, headers map[string]string) (*httpclient.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Put(rest.makeUrl(url), body, headers)
}
func (rest *RestClient2) Delete(url string, headers map[string]string) (*httpclient.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Delete(rest.makeUrl(url), headers)
}
func (rest *RestClient2) Patch(url string, query url.Values, body []byte, headers map[string]string) (*httpclient.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Patch(rest.makeUrl(url), query, body, headers)
}

func (rest *RestClient2) GetResponseRequstId(resp *httpclient.Response) string {
	return resp.Headers.Get("X-Openstack-Request-Id")
}

func NewRestClient2(baseUrl string, authPlugin auth.AuthPlugin) RestClient2 {
	return RestClient2{
		BaseUrl: baseUrl,
		session: &httpclient.Client{
			AuthPlugin:  authPlugin,
			BaseHeaders: map[string]string{"Content-Type": "application/json"},
		},
	}
}

type Openstack struct {
	keystoneClient *KeystoneV3
	glanceClient   *Glance
	neutronClient  *NeutronV2
	cinderClient   *CinderV2
	novaClient     *NovaV2
	AuthPlugin     auth.AuthPlugin
}

func (o Openstack) Region() string {
	return o.AuthPlugin.Region()
}

func NewClient(authUrl string, user auth.User, project auth.Project, regionName string) *Openstack {
	passwordAuth := auth.NewPasswordAuth(authUrl, user, project, regionName)
	return &Openstack{AuthPlugin: &passwordAuth}
}

func Client(region string) *Openstack {
	user := auth.User{
		Name:     common.CONF.Auth.User.Name,
		Password: common.CONF.Auth.User.Password,
		Domain:   auth.Domain{Name: common.CONF.Auth.User.Domain.Name},
	}
	project := auth.Project{
		Name: common.CONF.Auth.Project.Name,
		Domain: auth.Domain{
			Name: common.CONF.Auth.Project.Domain.Name,
		},
	}
	c := NewClient(common.CONF.Auth.Url, user, project, region)
	c.AuthPlugin.SetLocalTokenExpire(common.CONF.Auth.TokenExpireTime)
	return c
}

func DefaultClient() *Openstack {
	return Client(common.CONF.Auth.Region.Id)
}
