package identity

type Catalog struct {
	Type      string     `json:"type"`
	Name      string     `json:"name"`
	Id        string     `json:"id"`
	Endpoints []Endpoint `json:"endpoints"`
}

type RespToken struct {
	Token         Token `json:"token"`
	XSubjectToken string
}

type OptionCatalog struct {
	Region    string
	Type      string
	Name      string
	Interface string
}

type HttpException struct {
	Status  int
	Message string
}
