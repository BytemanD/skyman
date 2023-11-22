package identity

type Domain struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type User struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	Project     string `json:"project,omitempty"`
	Description string `json:"description,omitempty"`
	Email       string `json:"email,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	Domain      Domain `json:"domain"`
	DomainId    string `json:"domain_id,omitempty"`
}

type Password struct {
	User User `json:"user"`
}

type Identity struct {
	Methods  []string `json:"methods,omitempty"`
	Password Password `json:"password,omitempty"`
}

type Project struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Domain      Domain `json:"domain,omitempty"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled,omitempty"`
	DomainId    string `json:"domain_id,omitempty"`
}
type Scope struct {
	Project Project `json:"project,omitempty"`
}
type Auth struct {
	Identity Identity `json:"identity,omitempty"`
	Scope    Scope    `json:"scope,omitempty"`
}

type AuthBody struct {
	Auth Auth `json:"auth"`
}

type RoleAssigment struct {
	Scope Scope `json:"scope,omitempty"`
	User  User  `json:"user,omitempty"`
}

type Region struct {
	Name string `json:"name,omitempty"`
}
