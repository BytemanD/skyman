package common

type Resource struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
	ProjectId   string `json:"project_id,omitempty"`
	TenantId    string `json:"tenant_id,omitempty"`
	UserId      string `json:"user_id,omitempty"`
}
