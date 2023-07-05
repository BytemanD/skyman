package identity

type Domain struct {
	Name string `json:"name"`
}

type User struct {
	Domain   Domain `json:"domain"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type Password struct {
	User User `json:"user"`
}

type Identity struct {
	Methods  []string `json:"methods"`
	Password Password `json:"password"`
}

type Project struct {
	Domain Domain `json:"domain"`
	Name   string `json:"name"`
}
type Scope struct {
	Project Project `json:"project"`
}
type Auth struct {
	Identity Identity `json:"identity"`
	Scope    Scope    `json:"scope"`
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
