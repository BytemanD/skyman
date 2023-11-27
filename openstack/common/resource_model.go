package common

import (
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
