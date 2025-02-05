/*
OpenStack Client with Golang
*/
package openstack

import (
	"errors"
	"sync"
	"time"

	"github.com/BytemanD/go-console/console"

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
	region            string
	neutronEndpoint   string

	servieLock *sync.Mutex

	novaClient     *internal.NovaV2
	keystoneClient *internal.KeystoneV3
	glanceClient   *internal.GlanceV2
	cinderClient   *internal.CinderV2
	neutronClient  *internal.NeutronV2
}

func (o Openstack) Region() string {
	return utility.OneOfString(o.region, "RegionOne")
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
	}
}

func (o *Openstack) GlanceV2() *internal.GlanceV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.glanceClient == nil {
		endpoint, err := o.AuthPlugin.GetEndpoint(o.Region(), IMAGE, GLANCE, PUBLIC)
		if err != nil {
			console.Fatal("%s", errors.Unwrap(err))
		}
		o.glanceClient = &internal.GlanceV2{
			ServiceClient: internal.NewServiceApi(endpoint, V2, o.AuthPlugin),
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
		endpoint, err = o.AuthPlugin.GetEndpoint(o.Region(), VOLUME_V2, CINDER_V2, PUBLIC)
		if err != nil {
			console.Fatal("get cinder endpoint falied: %v", err)

		}
		o.cinderClient = &internal.CinderV2{
			ServiceClient: internal.NewServiceApi(endpoint, V2, o.AuthPlugin),
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
			endpoint, err = o.AuthPlugin.GetEndpoint(o.Region(), NETWORK, NEUTRON, PUBLIC)
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
		endpoint, err := o.AuthPlugin.GetEndpoint(o.Region(), IDENTITY, KEYSTONE, PUBLIC)
		if err != nil {
			console.Fatal("%v", err)
		}
		o.keystoneClient = &internal.KeystoneV3{
			ServiceClient: internal.NewServiceApi(endpoint, V3, o.AuthPlugin),
		}
	}
	return o.keystoneClient
}
func (o *Openstack) NovaV2(microVersion ...string) *internal.NovaV2 {
	o.servieLock.Lock()
	defer o.servieLock.Unlock()

	if o.novaClient == nil {
		endpoint, err := o.AuthPlugin.GetEndpoint(o.Region(), COMPUTE, NOVA, PUBLIC)
		if err != nil {
			console.Warn("get nova endpoint falied: %v", err)
			return &internal.NovaV2{
				ServiceClient: internal.NewServiceApi(endpoint, V2_1, o.AuthPlugin),
			}
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
