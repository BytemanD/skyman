package internal

import (
	"fmt"
	"strings"
)

type UrlPath string

func (t UrlPath) WithDetail() UrlPath {
	return UrlPath(strings.Join([]string{string(t), "detail"}, "/"))
}

// 添加路径
func (t UrlPath) A(path string) UrlPath {
	return UrlPath(strings.Join([]string{string(t), path}, "/"))
}

// 格式化后的值
func (t UrlPath) F(args ...any) string {
	return fmt.Sprintf(string(t), args...)
}

const (
	// server
	URL_SERVERS             UrlPath = "servers"
	URL_SERVERS_DETAIL      UrlPath = "servers/detail"
	URL_SERVER              UrlPath = "servers/%s"
	URL_SERVER_VOLUMES_BOOT UrlPath = "os-volumes_boot"
	// URL_SERVER_METADATA = "servers/%s/metadata"
	// URL_SERVER_METADATA_ITEM = "servers/%s/metadata/%s"
	URL_SERVER_CONSOLE         UrlPath = "servers/%s/console"
	URL_COMPUTE_SERVICES       UrlPath = "os-services"
	URL_COMPUTE_SERVICE        UrlPath = "os-services/%s"
	URL_COMPUTE_SERVICE_ACTION UrlPath = "os-services/%s"
	URL_HYPERVISORS            UrlPath = "os-hypervisors"
	URL_HYPERVISORS_DETAIL     UrlPath = "os-hypervisors/detail"
	URL_HYPERVISOR             UrlPath = "os-hypervisors/%s"
	URL_HYPERVISOR_UPTIME      UrlPath = "os-hypervisors/%s/uptime"
	URL_HYPERVISOR_CAPACITIES  UrlPath = "os-hypervisors/statistics/flavor-capacities"
	URL_SERVER_ACTION          UrlPath = "servers/%s/action"
	// 虚拟机接口
	URL_SERVER_INTERFACES UrlPath = "servers/%s/os-interface"
	URL_SERVER_INTERFACE  UrlPath = "servers/%s/os-interface/%s"
	// 安全组
	URL_SERVER_SECURITY_GROUPS UrlPath = "servers/%s/os-security-groups"
	URL_SERVER_SECURITY_GROUP  UrlPath = "servers/%s/os-security-groups/%s"
	// 密钥对
	URL_KEYPAIRS UrlPath = "os-keypairs"
	URL_KEYPAIR  UrlPath = "os-keypairs/%s"
	// 可用区
	URL_AVAILABILITY_ZONES        UrlPath = "os-availability-zone"
	URL_AVAILABILITY_ZONES_DETAIL UrlPath = "os-availability-zone/detail"
	// 聚合
	URL_AGGREGATES        UrlPath = "os-aggregates"
	URL_AGGREGATE         UrlPath = "os-aggregates/%s"
	URL_AGGREGATE_ACTION  UrlPath = "os-aggregates/%s/action"
	URL_AGGREGATES_DETAIL UrlPath = "os-aggregates/detail"
	// 虚拟机迁移
	URL_MIGRATIONS_LIST UrlPath = "os-migrations"
	// 配额
	URL_QUOTAS_LIST    UrlPath = "os-quota-sets"
	URL_QUOTA_DETAIL   UrlPath = "os-quota-sets/%s"
	URL_QUOTA_DEFAULTS UrlPath = "os-quota-sets/%s/defaults"
	// 虚拟机类型
	URL_FLAVORS            UrlPath = "flavors"
	URL_FLAVORS_DETAIL     UrlPath = "flavors/detail"
	URL_FLAVOR             UrlPath = "flavors/%s"
	URL_FLAVOR_EXTRA_SPECS UrlPath = "flavors/%s/os-extra_specs"
	URL_FLAVOR_EXTRA_SPEC  UrlPath = "flavors/%s/os-extra_specs/%s"
	// 虚拟机卷
	URL_SERVER_VOLUMES UrlPath = "servers/%s/os-volume_attachments"
	URL_SERVER_VOLUME  UrlPath = "servers/%s/os-volume_attachments/%s"
	//
	URL_SERVERS_REMOTE_CONSOLES UrlPath = "servers/%s/remote-consoles"
	//
	URL_SERVER_INSTANCE_ACTIONS UrlPath = "servers/%s/os-instance-actions"
	URL_SERVER_INSTANCE_ACTION  UrlPath = "servers/%s/os-instance-actions/%s"
	URL_SERVER_MIGRATIONS       UrlPath = "servers/%s/migrations"
	// 群组
	URL_SERVER_GROUPS UrlPath = "os-server-groups"
	// 虚拟机标签
	URL_SERVER_TAGS = "servers/%s/tags"
	URL_SERVER_TAG  = "servers/%s/tags/%s"

	// cinder

	// 卷
	URL_VOLUMES        UrlPath = "volumes"
	URL_VOLUMES_DETAIL UrlPath = "volumes/detail"
	URL_VOLUME         UrlPath = "volumes/%s"
	URL_VOLUME_ACTION  UrlPath = "volumes/%s/action"
	// 快照
	URL_SNAPSHOTS        UrlPath = "snapshots"
	URL_SNAPSHOTS_DETAIL UrlPath = "snapshots/detail"
	URL_SNAPSHOT         UrlPath = "snapshots/%s"
	// 备份
	URL_BACKUPS        UrlPath = "backups"
	URL_BACKUPS_DETAIL UrlPath = "backups/detail"
	URL_BACKUP         UrlPath = "backups/%s"
	// 卷类型
	URL_VOLUME_TYPES        UrlPath = "types"
	URL_VOLUME_TYPE         UrlPath = "types/%s"
	URL_VOLUME_TYPE_DEFAULT UrlPath = "types/default"
	URL_VOLUME_SERVICES     UrlPath = "os-services"

	// glance

	// 镜像
	URL_IMAGES     UrlPath = "images"
	URL_IMAGE      UrlPath = "images/%s"
	URL_IMAGE_FILE UrlPath = "images/%s/file"

	// keystone

	// region
	URL_REGIONS UrlPath = "regions"
	// 服务
	URL_SERVICES UrlPath = "services"
	URL_SERVICE  UrlPath = "services/%s"
	// endpoint
	URL_ENDPOINTS UrlPath = "endpoints"
	URL_ENDPOINT  UrlPath = "endpoints/%s"
	// 租户
	URL_PROJECTS UrlPath = "projects"
	URL_PROJECT  UrlPath = "projects/%s"
	// 用户
	URL_USERS UrlPath = "users"
	URL_USER  UrlPath = "users/%s"
	// 角色
	URL_ROLE_ASSIGNMENTS UrlPath = "role_assignments"

	// neutron

	URL_AGENTS UrlPath = "agents"

	URL_NETWORKS UrlPath = "networks"
	URL_NETWORK  UrlPath = "networks/%s"

	URL_SUBNETS UrlPath = "subnets"
	URL_SUBNET  UrlPath = "subnets/%s"

	URL_PORTS UrlPath = "ports"
	URL_PORT  UrlPath = "ports/%s"

	URL_ROUTERS                 UrlPath = "routers"
	URL_ROUTER                  UrlPath = "routers/%s"
	URL_ROUTER_ADD_INTERFACE    UrlPath = "routers/%s/add_router_interface"
	URL_ROUTER_REMOVE_INTERFACE UrlPath = "routers/%s/remove_router_interface"

	URL_FLOATINGIPS UrlPath = "floatingips"
	URL_FLOATINGIP  UrlPath = "floatingips/%s"

	URL_SECURITY_GROUPS      UrlPath = "security-groups"
	URL_SECURITY_GROUP       UrlPath = "security-groups/%s"
	URL_SECURITY_GROUP_RULES UrlPath = "security-group-rules"
	URL_SECURITY_GROUP_RULE  UrlPath = "security-group-rules/%s"

	URL_LOADBALANCERS UrlPath = "lbaas/loadbalancers"
	URL_LOADBALANCER  UrlPath = "lbaas/loadbalancers/%s"
	URL_LISTENERS     UrlPath = "lbaas/listeners"
	URL_LISTENER      UrlPath = "lbaas/listeners/%s"

	URL_POOLS UrlPath = "lbaas/pools"
	URL_POOL  UrlPath = "lbaas/pools/%s"

	URL_HEALTH_MONITORS UrlPath = "lbaas/healthmonitors"
	URL_HEALTH_MONITOR  UrlPath = "lbaas/healthmonitors/%s"

	URL_FIREWALLS         UrlPath = "fw/firewalls"
	URL_FIREWALL          UrlPath = "fw/firewalls/%s"
	URL_FIREWALL_POLICIES UrlPath = "fw/firewall_policies"
	URL_FIREWALL_POLICY   UrlPath = "fw/firewall_policies/%s"
	URL_FIREWALL_RULES    UrlPath = "fw/firewall_rules"
	URL_FIREWALL_RULE     UrlPath = "fw/firewall_rules/%s"

	URL_QOS_POLICIES     UrlPath = "qos/policies"
	URL_QOS_POLICY       UrlPath = "qos/policies/%s"
	URL_QOS_POLICY_RULES UrlPath = "qos/policies-rules"
)

// response body key
const (
	SERVICES = "services"
	SERVICE  = "service"
	PROJECTS = "projects"
	PROJECT  = "project"
	USERS    = "users"
	USER     = "user"

	VOLUMES   = "volumes"
	SNAPSHOTS = "snapshots"
	BACKUPS   = "backups"

	ROUTERS  = "routers"
	ROUTER   = "router"
	NETWORKS = "networks"
	NETWORK  = "network"
	SUBNETS  = "subnets"
	SUBNET   = "subnet"
	PORTS    = "ports"
	PORT     = "port"

	SERVERS = "servers"
	SERVER  = "server"
)
