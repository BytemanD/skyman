package identity

type Domain struct {
	Id   string `json:"id"`
	Name string `json:"name"`
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
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
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

func GetAuthReqBody(username string, password string, project_name string) AuthBody {
	authBody := AuthBody{}
	authBody.Auth.Identity.Methods = []string{"password"}

	authBody.Auth.Identity.Password.User.Name = username
	authBody.Auth.Identity.Password.User.Password = password
	authBody.Auth.Identity.Password.User.Domain.Name = "default"
	authBody.Auth.Scope.Project.Name = project_name
	authBody.Auth.Scope.Project.Domain.Name = "default"

	return authBody
}

type RoleAssigment struct {
	Scope Scope `json:"scope,omitempty"`
	User  User  `json:"user,omitempty"`
}
