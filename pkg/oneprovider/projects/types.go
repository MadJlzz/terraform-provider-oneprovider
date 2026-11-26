package projects

type ProjectReadResponse struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	IsDefault int    `json:"is_default"`
}

type ProjectCreateResponse struct {
	Response struct {
		Project ProjectReadResponse `json:"project"`
	} `json:"response"`
}

type ProjectsListResponse struct {
	Response struct {
		Projects []ProjectReadResponse `json:"projects"`
	} `json:"response"`
}
