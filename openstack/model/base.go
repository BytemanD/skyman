package model

import (
	"fmt"
	"reflect"
	"strings"
)

type Resource struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
	ProjectId   string `json:"project_id,omitempty"`
	TenantId    string `json:"tenant_id,omitempty"`
	UserId      string `json:"user_id,omitempty"`
}

func (resource Resource) GetStructTags() []string {
	tags := []string{}
	iType := reflect.TypeOf(resource)
	for i := 0; i < iType.NumField(); i++ {
		tag := iType.Field(i).Tag
		values := strings.Split(tag.Get("json"), ",")
		if len(values) >= 1 {
			tags = append(tags, strings.TrimSpace(values[0]))
		}
	}
	return tags
}
func (resource Resource) NameOrId() string {
	if resource.Name != "" {
		return resource.Name
	} else {
		return resource.Id
	}
}
func (resource Resource) IsActive() bool {
	return strings.EqualFold(resource.Status, "ACTIVE")
}
func (resource Resource) IsError() bool {
	return strings.EqualFold(resource.Status, "ERROR")
}

type ApiVersion struct {
	Id         string `json:"id"`
	MinVersion string `json:"min_version"`
	Status     string `json:"status"`
	Updated    string `json:"updated"`
	Version    string `json:"version"`
}

func (v *ApiVersion) VersoinInfo() string {
	version := v.Id
	if v.MinVersion != "" {
		version += fmt.Sprintf(" (%s ~ %s)", v.MinVersion, v.Version)
	}
	return version
}

type ApiVersions []ApiVersion

func (client ApiVersions) Current() *ApiVersion {
	for _, version := range client {
		if strings.EqualFold(version.Status, "CURRENT") {
			return &version
		}
	}
	return nil
}

func (client ApiVersions) Stable() *ApiVersion {
	for _, version := range client {
		if strings.EqualFold(version.Status, "STABLE") {
			return &version
		}
	}
	return nil
}

type RequestId struct {
	requestId string
}

func (r *RequestId) SetRequestId(requestId string) {
	r.requestId = requestId
}
func (r RequestId) GetRequestId() string {
	return r.requestId
}
