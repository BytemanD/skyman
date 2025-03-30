/*
OpenStack Client with Golang
*/
package openstack

import (
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
			MicroVersion: &model.ApiVersion{Version: "2.1"},
		}
		if err := o.novaClient.DiscoverMicroVersion(); err != nil {
			console.Warn("get current version failed: %v", err)
		}
		if o.novaClient.MicroVersion != nil {
			// o.novaClient.SetHeader("Openstack-Api-Version", o.novaClient.MicroVersion.Version)
			o.novaClient.SetHeader("X-Openstack-Nova-Api-Version", o.novaClient.MicroVersion.Version)
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
