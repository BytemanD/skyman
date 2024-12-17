/*
OpenStack Client with Golang
*/
package openstack

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/BytemanD/easygo/pkg/global/logging"
	"github.com/BytemanD/skyman/common"
	"github.com/BytemanD/skyman/openstack/auth"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/utility"
	"github.com/BytemanD/skyman/utility/httpclient"
	"github.com/go-resty/resty/v2"
)

const X_OPENSTACK_REQUEST_ID = "X-Openstack-Request-Id"

type RestClient2 struct {
	BaseUrl string
	session *httpclient.RESTClient
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
	if rest.session == nil && rest.session.BaseHeaders == nil {
		return
	}
	rest.session.BaseHeaders[key] = value
}
func (rest *RestClient2) Index() (*resty.Response, error) {
	parsed, err := url.Parse(rest.BaseUrl)
	if err != nil {
		return nil, err
	}
	return rest.session.Get(
		fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host), nil, nil)
}
func (rest *RestClient2) Get(url string, query url.Values) (*resty.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Get(rest.makeUrl(url), query, nil)
}
func (rest *RestClient2) GetAndUnmarshal(url string, query url.Values, body interface{}) error {
	resp, err := rest.Get(url, query)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp.Body(), &body); err != nil {
		return err
	}
	return nil
}

func (rest *RestClient2) Post(url string, body interface{}, headers map[string]string) (*resty.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Post(rest.makeUrl(url), body, headers)
}
func (rest *RestClient2) Put(url string, body interface{}, headers map[string]string) (*resty.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Put(rest.makeUrl(url), body, headers)
}
func (rest *RestClient2) Delete(url string, headers map[string]string) (*resty.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Delete(rest.makeUrl(url), headers)
}
func (rest *RestClient2) Patch(url string, query url.Values, body interface{}, headers map[string]string) (*resty.Response, error) {
	if err := rest.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return rest.session.Patch(rest.makeUrl(url), body, query, headers)
}

func (rest *RestClient2) GetResponseRequstId(resp *resty.Response) string {
	return resp.Header().Get(X_OPENSTACK_REQUEST_ID)
}

func NewRestClient2(baseUrl string, authPlugin httpclient.AuthPluginInterface) RestClient2 {
	client := RestClient2{
		BaseUrl: baseUrl,
		session: httpclient.New().SetAuthPlugin(authPlugin),
	}
	return client
}

type Openstack struct {
	keystoneClient    *KeystoneV3
	glanceClient      *Glance
	neutronClient     *NeutronV2
	cinderClient      *CinderV2
	novaClient        *NovaV2
	AuthPlugin        auth.AuthPlugin
	ComputeApiVersion string
	servieLock        *sync.Mutex
}

func (o Openstack) Region() string {
	return o.AuthPlugin.Region()
}
func (o Openstack) ProjectId() (string, error) {
	return o.AuthPlugin.GetProjectId()
}

func NewClient(authUrl string, user auth.User, project auth.Project, regionName string) *Openstack {
	authUrl = utility.VersionUrl(authUrl, fmt.Sprintf("v%s", common.CONF.Identity.Api.Version))
	passwordAuth := auth.NewPasswordAuth(authUrl, user, project, regionName)
	logging.Debug("new opensack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		common.CONF.HttpTimeoutSecond, common.CONF.RetryWaitTimeSecond, common.CONF.RetryCount,
	)
	passwordAuth.SetHttpTimeout(common.CONF.HttpTimeoutSecond)
	passwordAuth.SetRetryWaitTime(common.CONF.RetryWaitTimeSecond)
	passwordAuth.SetRetryCount(common.CONF.RetryCount)

	return &Openstack{AuthPlugin: passwordAuth, servieLock: &sync.Mutex{}}
}

func ClientWithRegion(region string) *Openstack {
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
	c := ClientWithRegion(common.CONF.Auth.Region.Id)
	c.ComputeApiVersion = "2.1"
	return c
}

func getMicroVersion(vertionStr string) microVersion {
	versionList := strings.Split(vertionStr, ".")
	v, _ := strconv.Atoi(versionList[0])
	micro, _ := strconv.Atoi(versionList[1])
	return microVersion{Version: v, MicroVersion: micro}
}

type ResourceApi struct {
	Endpoint        string
	BaseUrl         string
	Client          *httpclient.RESTClient
	MicroVersion    model.ApiVersion
	EnableAllTenant bool
	query           url.Values
	headers         map[string]string
	body            interface{}
	result          interface{}
	SingularKey     string
	PluralKey       string
}

func (api ResourceApi) makeUrl() string {
	result, _ := url.JoinPath(api.Endpoint, api.BaseUrl)
	return result
}
func (api ResourceApi) mustHasBaseUrl() error {
	if api.BaseUrl == "" {
		return fmt.Errorf("base url is required")
	}
	return nil
}
func (api *ResourceApi) MicroVersionLargeEqual(version string) bool {
	clientVersion := api.GetMicroVersion()
	return clientVersion.LargeEqual(version)
}
func (api *ResourceApi) GetMicroVersion() microVersion {
	return getMicroVersion(api.MicroVersion.Version)
}
func (api *ResourceApi) SetHeader(h, v string) *ResourceApi {
	if api.headers == nil {
		api.headers = map[string]string{}
	}
	api.headers[h] = v
	return api
}
func (api *ResourceApi) SetHeaders(headers map[string]string) *ResourceApi {
	if api.headers == nil {
		api.headers = map[string]string{}
	}
	for h, v := range headers {
		api.headers[h] = v
	}
	return api
}
func (api *ResourceApi) SetQuery(query url.Values) *ResourceApi {
	api.query = query
	return api
}
func (api *ResourceApi) AddQuery(k, v string) *ResourceApi {
	if api.query == nil {
		api.query = url.Values{}
	}
	api.query.Set(k, v)
	return api
}
func (api *ResourceApi) SetBody(body interface{}) *ResourceApi {
	api.body = body
	return api
}
func (api *ResourceApi) SetResult(result interface{}) *ResourceApi {
	api.result = result
	return api
}
func (api *ResourceApi) AppendUrl(url string) *ResourceApi {
	api.BaseUrl = utility.UrlJoin(api.BaseUrl, url)
	return api
}
func (api *ResourceApi) AppendUrlf(u string, args ...any) *ResourceApi {
	api.BaseUrl = utility.UrlJoin(api.BaseUrl, fmt.Sprintf(u, args...))
	return api
}
func (api *ResourceApi) PopUrl() *ResourceApi {
	if api.BaseUrl != "" {
		values := strings.Split(api.BaseUrl, "/")
		api.BaseUrl = utility.UrlJoin(values[0 : len(values)-1]...)
	}
	return api
}
func (api *ResourceApi) SetUrl(url string) *ResourceApi {
	api.BaseUrl = url
	return api
}
func (api *ResourceApi) Index() (*resty.Response, error) {
	if err := api.mustHasBaseUrl(); err != nil {
		return nil, err
	}
	return api.Client.Get(utility.UrlJoin(api.Endpoint, api.BaseUrl), nil, nil)
}

func (api ResourceApi) Get(res interface{}) (*resty.Response, error) {
	resp, err := api.Client.Get(api.makeUrl(), api.query, api.headers)
	if err != nil || res == nil {
		return resp, err
	}
	return resp, json.Unmarshal(resp.Body(), res)
}

func (api ResourceApi) Post(res interface{}) (*resty.Response, error) {
	resp, err := api.Client.Post(api.makeUrl(), api.body, api.headers)
	if err != nil || res == nil {
		return resp, err
	}
	return resp, json.Unmarshal(resp.Body(), res)
}

func (api ResourceApi) Put(res interface{}) (*resty.Response, error) {
	resp, err := api.Client.Put(api.makeUrl(), api.body, api.headers)
	if err != nil || res == nil {
		return resp, err
	}
	return resp, json.Unmarshal(resp.Body(), res)
}
func (api ResourceApi) Delete(res interface{}) (*resty.Response, error) {
	resp, err := api.Client.Delete(api.makeUrl(), api.headers)
	if err != nil || res == nil {
		return resp, err
	}
	return resp, json.Unmarshal(resp.Body(), res)
}
func (api ResourceApi) Patch(res interface{}) (*resty.Response, error) {
	resp, err := api.Client.Patch(api.makeUrl(), api.body, api.query, api.headers)
	if err != nil || res == nil {
		return resp, err
	}
	return resp, json.Unmarshal(resp.Body(), res)
}
func (api ResourceApi) IsAdmin() bool {
	return api.Client.AuthPlugin.IsAdmin()
}

type ResourceInterface interface {
	GetName() string
}

func FoundResource[T any](api ResourceApi, idOrName string) (*T, error) {
	if api.SingularKey == "" || api.PluralKey == "" {
		return nil, fmt.Errorf("resource api %v SingularKey or PluralKey is empty", api)
	}
	resp, err := api.AppendUrl(idOrName).Get(nil)
	if err == nil {
		body := map[string]*T{}
		if err := json.Unmarshal(resp.Body(), &body); err != nil {
			return nil, err
		}
		return body[api.SingularKey], nil
	}
	if _, ok := err.(httpclient.HttpError); !ok {
		return nil, err
	}
	if httpError, _ := err.(httpclient.HttpError); !httpError.IsNotFound() {
		return nil, err
	}
	api.PopUrl().AddQuery("name", idOrName)
	if api.IsAdmin() && api.EnableAllTenant {
		api.AddQuery("all_tenants", "true")
	}

	resp, err = api.Get(nil)
	if err != nil {
		return nil, err
	}

	body2 := []T{}
	if err := utility.UnmarshalJsonKey(resp.Body(), api.PluralKey, &body2); err != nil {
		return nil, err
	}
	switch len(body2) {
	case 0:
		return nil, fmt.Errorf("resource %s not found", idOrName)
	case 1:
		t := body2[0]
		value := reflect.ValueOf(t)
		valueName := value.FieldByName("Name")

		if valueName.String() != idOrName {
			return nil, fmt.Errorf("resource %s not found", idOrName)
		} else {
			return &t, nil
		}
	default:
		fileted := []T{}
		for _, t := range body2 {
			value := reflect.ValueOf(t)
			valueName := value.FieldByName("Name")
			if valueName.Kind() == reflect.String && valueName.String() == idOrName {
				fileted = append(fileted, t)
			}
		}
		if len(fileted) == 0 {
			return nil, fmt.Errorf("resource %s not found", idOrName)
		}
		if len(fileted) == 1 {
			return &fileted[0], nil
		}
		return nil, fmt.Errorf("found %d resources with name %s ", len(fileted), idOrName)
	}
}
