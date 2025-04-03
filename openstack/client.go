/*
OpenStack Client with Golang
*/
package openstack

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/BytemanD/go-console/console"
	"github.com/samber/lo"

	"github.com/BytemanD/skyman/openstack/internal"
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/model"
)

const (
	V2   = "v2"
	V2_0 = "v2.0"
	V2_1 = "v2.1"
	V3   = "v3"

	IDENTITY  = "identity"
	NETWORK   = "network"
	VOLUME    = "volume"
	VOLUME_V2 = "volumev2"
	VOLUME_V3 = "volumev3"
	STORAGE   = "storage"
	COMPUTE   = "compute"
	IMAGE     = "image"

	KEYSTONE  = "keystone"
	NOVA      = "nova"
	CINDER    = "cinder"
	CINDER_V2 = "cinderv2"
	CINDER_V3 = "cinderv3"
	GLANCE    = "glance"
	NEUTRON   = "neutron"

	PUBLIC   = "public"
	INTERNAL = "internal"
	ADMIN    = "admin"
)

var COMPUTE_API_VERSION string

type Openstack struct {
	AuthPlugin        auth_plugin.AuthPlugin
	ComputeApiVersion string
	region            string
	neutronEndpoint   string

	servieLock *sync.Mutex

	novaClient     *internal.NovaV2
	keystoneClient *internal.KeystoneV3
	glanceClient   *internal.GlanceV2
	cinderClient   *internal.CinderV2
	neutronClient  *internal.NeutronV2

	novaClientOnce   *sync.Once
	cinderClientOnce *sync.Once

	cloudConfig Cloud
}

func (o *Openstack) IsAdmin() bool {
	return o.AuthPlugin != nil && o.AuthPlugin.IsAdmin()
}
func (o Openstack) Region() string {
	return lo.CoalesceOrEmpty(o.region, "RegionOne")
}
func (o *Openstack) ResetAllClients() {
	o.keystoneClient = nil
	o.glanceClient = nil
	o.cinderClient = nil
	o.novaClient = nil
	o.neutronClient = nil
}

func (o *Openstack) SetRegion(region string) *Openstack {
	if o.Region() == region {
		return o
	}
	o.region = region
	o.ResetAllClients()
	return o
}
func (o *Openstack) WithRegion(region string) *Openstack {
	return &Openstack{
		AuthPlugin:        o.AuthPlugin,
		ComputeApiVersion: o.ComputeApiVersion,
		region:            region,
		neutronEndpoint:   o.neutronEndpoint,

		servieLock: &sync.Mutex{},
	}
}
func (o Openstack) ProjectId() (string, error) {
	return o.AuthPlugin.GetProjectId()
}
func (o *Openstack) SetNeutronEndpoint(endpoint string) {
	o.neutronEndpoint = endpoint
}
func (o *Openstack) SetComputeApiVersion(version string) {
	o.ComputeApiVersion = version
}

func NewClient(authUrl string, user model.User, project model.Project, regionName string) *Openstack {
	return &Openstack{
		AuthPlugin:        internal.NewPasswordAuth(authUrl, user, project),
		ComputeApiVersion: COMPUTE_API_VERSION,
		region:            regionName,
		servieLock:        &sync.Mutex{},
		novaClientOnce:    &sync.Once{},
		cinderClientOnce:  &sync.Once{},
	}
}

func (o *Openstack) GlanceV2() *internal.GlanceV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.glanceClient == nil {
		o.glanceClient = &internal.GlanceV2{
			ServiceClient: internal.NewServiceClient(
				o.Region(), IMAGE, GLANCE, PUBLIC, V2_1, o.AuthPlugin,
			),
		}
	}
	return o.glanceClient
}

func (o *Openstack) CinderV2() *internal.CinderV2 {
	o.cinderClientOnce.Do(func() {
		o.cinderClient = &internal.CinderV2{
			ServiceClient: internal.NewServiceClient(
				o.Region(), VOLUME_V2, CINDER_V2, PUBLIC, V2, o.AuthPlugin,
			),
		}
	})
	return o.cinderClient
}

func (o *Openstack) NeutronV2() *internal.NeutronV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.neutronClient == nil {
		o.neutronClient = &internal.NeutronV2{
			ServiceClient: internal.NewServiceClient(
				o.Region(), NETWORK, NEUTRON, PUBLIC, V2_0, o.AuthPlugin,
			),
		}
		o.neutronClient.Client.BaseURL = o.cloudConfig.Neutron.Endpoint
	}
	return o.neutronClient
}
func (o *Openstack) KeystoneV3() *internal.KeystoneV3 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.keystoneClient == nil {
		o.keystoneClient = &internal.KeystoneV3{
			ServiceClient: internal.NewServiceClient(
				o.Region(), IDENTITY, KEYSTONE, PUBLIC, V3, o.AuthPlugin,
			),
		}
	}
	return o.keystoneClient
}
func (o *Openstack) NovaV2(microVersion ...string) *internal.NovaV2 {
	o.novaClientOnce.Do(func() {
		o.novaClient = &internal.NovaV2{
			ServiceClient: internal.NewServiceClient(
				o.Region(), COMPUTE, NOVA, PUBLIC, V2_1, o.AuthPlugin,
			),
			// ApiVersion: model.ApiVersion{Version: "2.1"},
		}
		if o.cloudConfig.Compute.Api.Version != "" {
			// v := internal.ParsetVersionFromString(o.cloudConfig.Compute.Api.Version)
			// o.novaClient.ApiVersion = model.ApiVersion{
			// Version:    v.Version,
			// MinVersion: Version,
			// }
			o.novaClient.SetHeader(internal.X_OPENSTACK_NOVA_API_VERSION, o.cloudConfig.Compute.Api.Version)
		} else {
			if err := o.novaClient.DiscoverMicroVersion(); err != nil {
				console.Warn("get current version failed: %v", err)
			}
		}
	})
	return o.novaClient
}
func (o *Openstack) SetHttpTimeout(timeout time.Duration) {
	o.AuthPlugin.SetTimeout(timeout)
	if o.keystoneClient != nil {
		o.keystoneClient.Client.SetTimeout(timeout)
	}
}
func (o *Openstack) SetRetryWaitTime(timeout time.Duration) {
	o.AuthPlugin.SetRetryWaitTime(timeout)
	if o.keystoneClient != nil {
		o.keystoneClient.Client.SetRetryWaitTime(timeout)
	}
}
func (o *Openstack) SetRetryWaitMaxTime(timeout time.Duration) {
	o.AuthPlugin.SetRetryMaxWaitTime(timeout)
	if o.keystoneClient != nil {
		o.keystoneClient.Client.SetRetryMaxWaitTime(timeout)
	}
}
func (o *Openstack) SetRetryCount(count int) {
	o.AuthPlugin.SetRetryCount(count)
	if o.keystoneClient != nil {
		o.keystoneClient.Client.SetRetryCount(count)
	}
}

// export OS_REGION_NAME=RegionOne
// export OS_AUTH_URL=http://keystone.region1.dev:35357/v3
// export OS_USERNAME=admin
// export OS_PASSWORD=PASSWORD
// export OS_PROJECT_NAME=admin
// export OS_USER_DOMAIN_NAME=default
// export OS_PROJECT_DOMAIN_NAME=default
// export OS_IDENTITY_API_VERSION=3
// export OS_IMAGE_API_VERSION=2
// export OS_VOLUME_API_VERSION=2
// export OS_BAREMETAL_API_VERSION=latest
// export IRONIC_API_VERSION=latest
func loadFromEnv() {
	cloud.RegionName = lo.CoalesceOrEmpty(os.Getenv("OS_REGION_NAME"), cloud.RegionName)
	cloud.Auth.AuthUrl = lo.CoalesceOrEmpty(os.Getenv("OS_AUTH_URL"), cloud.Auth.AuthUrl)
	cloud.Auth.ProjectDomainId = lo.CoalesceOrEmpty(os.Getenv("OS_PROJECT_DOMAIN_NAME"), cloud.Auth.ProjectDomainId)
	cloud.Auth.UserDomainId = lo.CoalesceOrEmpty(os.Getenv("OS_USER_DOMAIN_NAME"), cloud.Auth.UserDomainId)
	cloud.Auth.ProjectName = lo.CoalesceOrEmpty(os.Getenv("OS_PROJECT_NAME"), cloud.Auth.ProjectName)
	cloud.Auth.Username = lo.CoalesceOrEmpty(os.Getenv("OS_USERNAME"), cloud.Auth.Username)
	cloud.Auth.Password = lo.CoalesceOrEmpty(os.Getenv("OS_PASSWORD"), cloud.Auth.Password)

	cloud.Identity.Api.Version = lo.CoalesceOrEmpty(os.Getenv("OS_IDENTITY_API_VERSION"), cloud.Identity.Api.Version)
	cloud.Neutron.Endpoint = lo.CoalesceOrEmpty(os.Getenv("OS_NEUTRON_ENDPOINT"), cloud.Neutron.Endpoint)
	cloud.Compute.Api.Version = lo.CoalesceOrEmpty(os.Getenv("OS_COMPUTE_API_VERSION"), cloud.Compute.Api.Version)

}
func connectCloud() (*Openstack, error) {
	// 更新默认配置
	cloud.TokenExpireTime = lo.CoalesceOrEmpty(cloud.TokenExpireTime, 60*30)

	conn := NewClient(
		cloud.Auth.AuthUrl,
		model.User{
			Name:     cloud.Auth.Username,
			Domain:   model.Domain{Name: cloud.Auth.UserDomainId},
			Password: cloud.Auth.Password,
		},
		model.Project{
			Name:   cloud.Auth.ProjectName,
			Domain: model.Domain{Name: cloud.Auth.ProjectDomainId},
		},
		cloud.Region(),
	)
	conn.cloudConfig = cloud
	conn.AuthPlugin.SetLocalTokenExpire(cloud.TokenExpireTime)
	if CONF.HttpTimeoutSecond > 0 {
		conn.SetHttpTimeout(time.Second * time.Duration(CONF.HttpTimeoutSecond))
	}
	if CONF.RetryWaitTimeSecond > 0 {
		conn.SetRetryWaitTime(time.Second * time.Duration(CONF.RetryWaitTimeSecond))
	}
	if CONF.RetryCount > 0 {
		conn.SetRetryCount(CONF.RetryCount)
	}
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		CONF.HttpTimeoutSecond, CONF.RetryWaitTimeSecond, CONF.RetryCount,
	)
	console.Debug("cloud: %v", cloud.Auth.ProjectDomainId)
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		CONF.HttpTimeoutSecond, CONF.RetryWaitTimeSecond, CONF.RetryCount,
	)
	if _, err := conn.AuthPlugin.GetToken(); err != nil {
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	return conn, nil
}
func GetOne(name string) (*Openstack, error) {
	if c, ok := CONF.Clouds[name]; !ok {
		return nil, fmt.Errorf("cloud %s not found", name)
	} else {
		cloud = c
		return connectCloud()
	}
}

// 如果指定 cloud 名称, 优先使用 cloud对应的配置;
// 否则从环境变量读取 cloud;
// 最后，使用默认的cloud.
func Connect(name ...string) (*Openstack, error) {
	useCloudName := lo.FirstOrEmpty(append(name, cloudName))
	if useCloudName != "" {
		return GetOne(useCloudName)
	}
	console.Debug("load cloud config from env")
	loadFromEnv()
	if cloud.Auth.AuthUrl == "" {
		return nil, fmt.Errorf("auth url is empty, forget to load env or set cloud name?")
	}
	u, err := url.Parse(cloud.Auth.AuthUrl)
	if err != nil {
		return nil, fmt.Errorf("parse auth url failed: %w", err)
	}
	if u.Path == "" || u.Path == "/" {
		u.Path = "v" + lo.CoalesceOrEmpty(cloud.Identity.Api.Version, "3")
		cloud.Auth.AuthUrl = u.String()
	}
	return connectCloud()
}
