package openstack

const (
	URL_DETAIL              = "detail"
	URL_SERVER_VOLUMES_BOOT = "os-volumes_boot"
	URL_VOLUME_ATTACH       = "%s/os-volume_attachments"
	URL_VOLUME_DETACH       = "%s/os-volume_attachments/%s"
	URL_INTERFACE_ATTACH    = "%s/os-interface"
	URL_INTERFACE_DETACH    = "%s/os-interface/%s"
)

var COMPUTE_API_VERSION string
