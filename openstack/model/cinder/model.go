package cinder

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"

	"github.com/BytemanD/skyman/openstack/model"
)

const (
	KB = 1024
	MB = KB * 1024
	GB = MB * 1024

	POLICY_NEVER     = "never"
	POLICY_ON_DEMAND = "on-demand"
)

var MIGRATION_POLICYS = []string{POLICY_NEVER, POLICY_ON_DEMAND}

func InvalidMIgrationPoicy(policy string) error {
	if !slice.Contain(MIGRATION_POLICYS, policy) {
		return fmt.Errorf("invalid migration policy: %s, supported: %s", policy, MIGRATION_POLICYS)
	}
	return nil
}

type Attachment struct {
	Id           string `json:"id,omitempty"`
	ServerId     string `json:"server_id,omitempty"`
	AttachmentId string `json:"attachment_id,omitempty"`
	Device       string `json:"device,omitempty"`
	AttachmentAt string `json:"attached_at,omitempty"`
	HostName     string `json:"host_name,omitempty"`
	VolumeId     string `json:"volume_id,omitempty"`
}
type Volume struct {
	model.Resource
	Size       uint   `json:"size,omitempty"`
	VolumeType string `json:"volume_type,omitempty"`
	Bootable   string `json:"bootable"`

	Attachments         []Attachment      `json:"attachments"`
	Metadata            map[string]string `json:"metadata"`
	AvailabilityZone    string            `json:"availability_zone"`
	Host                string            `json:"os-vol-host-attr:host"`
	Multiattach         bool              `json:"multiattach"`
	SourceVolid         string            `json:"source_volid"`
	GroupId             string            `json:"group_id"`
	TaskStatus          string            `json:"task_status"`
	VolumeImageMetadata map[string]string `json:"volume_image_metadata"`
	TenantId            string            `json:"os-vol-tenant-attr:tenant_id,omitempty"`
}

type Volumes []Volume

func (volume Volume) GetAttachmentList() []string {
	attachmentList := []string{}
	for _, attachment := range volume.Attachments {
		attachmentList = append(attachmentList,
			fmt.Sprintf("%s @ %s", attachment.Device, attachment.ServerId),
		)
	}
	return attachmentList
}
func (volume Volume) IsBootable() bool {
	return volume.Bootable == "true"
}

func (volume Volume) GetMetadataList() []string {
	metadataList := []string{}
	for k, v := range volume.Metadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}
func (volume Volume) GetImageMetadataList() []string {
	metadataList := []string{}
	for k, v := range volume.VolumeImageMetadata {
		metadataList = append(metadataList, fmt.Sprintf("%s=%s", k, v))
	}
	return metadataList
}
func (volume Volume) IsError() bool {
	return volume.Status == "error"
}
func (volume Volume) IsAvailable() bool {
	return volume.Status == "available"
}
func (volume Volume) IsInuse() bool {
	return volume.Status == "in-use"
}

type VolumeType struct {
	model.Resource
	QosSpecsId                 string            `json:"qos_specs_id,omitempty"`
	IsPublic                   bool              `json:"is_public"`
	IsEncrypted                bool              `json:"is_encrypted"`
	OsVolumeTypeAccessIsPublic bool              `json:"os-volume-type-access:is_public"`
	ExtraSpecs                 map[string]string `json:"extra_specs,omitempty"`
}

func (volumeType VolumeType) GetExtraSpecsList() []string {
	properties := []string{}
	for key, value := range volumeType.ExtraSpecs {
		properties = append(properties, fmt.Sprintf("%s=%s", key, value))
	}
	return properties
}

type Service struct {
	model.Resource
	Host           string `json:"host,omitempty"`
	Binary         string `json:"binary,omitempty"`
	Zone           string `json:"zone,omitempty"`
	Status         string `json:"status,omitempty"`
	State          string `json:"state,omitempty"`
	DisabledReason string `json:"disabled_reason,omitempty"`
}

type Snapshot struct {
	model.Resource
	Size      uint                   `json:"size,omitempty"`
	VolumeId  string                 `json:"volume_id,omitempty"`
	ProjectId string                 `json:"os-extended-snapshot-attributes:project_id,omitempty"`
	Progress  string                 `json:"os-extended-snapshot-attributes:progress,omitempty"`
	Metadata  map[string]interface{} `json:"metadata:progress,omitempty"`
}

type Backup struct {
	model.Resource
	Size      uint                   `json:"size,omitempty"`
	VolumeId  string                 `json:"volume_id,omitempty"`
	ProjectId string                 `json:"os-extended-snapshot-attributes:project_id,omitempty"`
	Progress  string                 `json:"os-extended-snapshot-attributes:progress,omitempty"`
	Metadata  map[string]interface{} `json:"metadata:progress,omitempty"`
}
