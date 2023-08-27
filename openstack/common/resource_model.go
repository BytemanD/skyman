package common

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
