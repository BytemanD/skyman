/*
OpenStack Client with Golang
*/
package openstack

import (
	"fmt"
	"sync"

	"github.com/BytemanD/go-console/console"
	"github.com/BytemanD/skyman/common"

	"github.com/BytemanD/skyman/openstack/internal"
	"github.com/BytemanD/skyman/openstack/internal/auth_plugin"
	"github.com/BytemanD/skyman/openstack/model"
	"github.com/BytemanD/skyman/utility"
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

	novaClient *internal.NovaV2

	keystoneClient *internal.KeystoneV3
	glanceClient   *internal.GlanceV2
	cinderClient   *internal.CinderV2
	neutronClient  *internal.NeutronV2

	servieLock *sync.Mutex

	neutronEndpoint string
}

func (o *Openstack) WithRegion(region string) *Openstack {
	if region == o.AuthPlugin.Region() {
		return o
	}
	authPlugin := o.AuthPlugin
	authPlugin.SetRegion(region)
	return &Openstack{
		AuthPlugin:        authPlugin,
		ComputeApiVersion: o.ComputeApiVersion,

		neutronEndpoint: o.neutronEndpoint,

		servieLock: &sync.Mutex{},
	}
}
func (o Openstack) Region() string {
	return o.AuthPlugin.Region()
}
func (o Openstack) ProjectId() (string, error) {
	return o.AuthPlugin.GetProjectId()
}

func NewClient(authUrl string, user model.User, project model.Project, regionName string) *Openstack {
	authUrl = utility.VersionUrl(authUrl, fmt.Sprintf("v%s", common.CONF.Identity.Api.Version))
	passwordAuth := internal.NewPasswordAuth(authUrl, user, project, regionName)
	console.Debug("new openstack client, HttpTimeoutSecond=%d RetryWaitTimeSecond=%d RetryCount=%d",
		common.CONF.HttpTimeoutSecond, common.CONF.RetryWaitTimeSecond, common.CONF.RetryCount,
	)
	passwordAuth.SetHttpTimeout(common.CONF.HttpTimeoutSecond)
	passwordAuth.SetRetryWaitTime(common.CONF.RetryWaitTimeSecond)
	passwordAuth.SetRetryCount(common.CONF.RetryCount)

	return &Openstack{AuthPlugin: passwordAuth, servieLock: &sync.Mutex{}}
}

func ClientWithRegion(region string) *Openstack {
	user := model.User{
		Name:     common.CONF.Auth.User.Name,
		Password: common.CONF.Auth.User.Password,
		Domain:   model.Domain{Name: common.CONF.Auth.User.Domain.Name},
	}
	project := model.Project{
		Name: common.CONF.Auth.Project.Name,
		Domain: model.Domain{
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
	c.neutronEndpoint = common.CONF.Neutron.Endpoint
	c.ComputeApiVersion = COMPUTE_API_VERSION
	return c
}

func (o *Openstack) GlanceV2() *internal.GlanceV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.glanceClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint(IMAGE, GLANCE, PUBLIC)
		if err != nil {
			console.Fatal("get glance endpoint falied: %v", err)

		}
		o.glanceClient = &internal.GlanceV2{
			ServiceClient: internal.NewServiceApi[internal.ServiceClient](endpoint, V2, o.AuthPlugin),
		}
	}
	return o.glanceClient
}

func (o *Openstack) CinderV2() *internal.CinderV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.cinderClient == nil {
		var (
			endpoint string
			err      error
		)
		endpoint, err = o.AuthPlugin.GetServiceEndpoint(VOLUME_V2, CINDER_V2, PUBLIC)
		if err != nil {
			console.Fatal("get cinder endpoint falied: %v", err)

		}
		o.cinderClient = &internal.CinderV2{
			ServiceClient: internal.NewServiceApi[internal.ServiceClient](endpoint, V2, o.AuthPlugin),
		}
	}
	return o.cinderClient
}

func (o *Openstack) NeutronV2() *internal.NeutronV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.neutronClient == nil {
		endpoint := o.neutronEndpoint
		if endpoint == "" {
			var err error
			endpoint, err = o.AuthPlugin.GetServiceEndpoint(NETWORK, NEUTRON, PUBLIC)
			if err != nil {
				console.Fatal("get neutron endpoint falied: %v", err)

			}
		}
		o.neutronClient = &internal.NeutronV2{
			ServiceClient: internal.NewServiceApi(endpoint, V2_0, o.AuthPlugin),
		}
	}
	return o.neutronClient
}
func (o *Openstack) KeystoneV3() *internal.KeystoneV3 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.keystoneClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint(IDENTITY, KEYSTONE, PUBLIC)
		if err != nil {
			console.Fatal("get keystone endpoint falied: %v", err)
		}
		o.keystoneClient = &internal.KeystoneV3{
			ServiceClient: internal.NewServiceApi[internal.ServiceClient](endpoint, V3, o.AuthPlugin),
		}
	}
	return o.keystoneClient
}
func (o *Openstack) NovaV2(microVersion ...string) *internal.NovaV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.novaClient == nil {
		endpoint, err := o.AuthPlugin.GetServiceEndpoint(COMPUTE, NOVA, PUBLIC)
		if err != nil {
			console.Warn("get nova endpoint falied: %v", err)

			return &internal.NovaV2{}
		}
		o.novaClient = &internal.NovaV2{
			ServiceClient: internal.NewServiceApi(endpoint, V2_1, o.AuthPlugin),
		}
		if o.ComputeApiVersion != "" {
			o.novaClient.MicroVersion = &model.ApiVersion{
				Version: o.ComputeApiVersion,
			}
		} else {
			console.Debug("get current version of nova")
			currentVersion, err := o.novaClient.GetCurrentVersion()
			if err != nil {
				console.Warn("get current version failed: %v", err)
				o.novaClient.MicroVersion = &model.ApiVersion{
					Version: V2_1,
				}
			} else {
				o.novaClient.MicroVersion = currentVersion
			}
		}
		console.Debug("current nova version: %s", o.novaClient.MicroVersion.VersoinInfo())
		if o.novaClient.MicroVersion != nil {
			o.novaClient.AddBaseHeader("Openstack-Api-Version", o.novaClient.MicroVersion.Version)
			o.novaClient.AddBaseHeader("X-Openstack-Nova-Api-Version", o.novaClient.MicroVersion.Version)
		}
	}
	return o.novaClient
}
