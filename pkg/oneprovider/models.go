package oneprovider

type ListVMTemplatesResponse struct {
	Templates []VMTemplateResponse `json:"response"`
}

type VMTemplateResponse struct {
	Id      int    `json:"id"`
	Name    string `json:"name"`
	Size    string `json:"size"`
	Display struct {
		Name        string `json:"name"`
		Display     string `json:"display"`
		Description string `json:"description"`
		Oca         int    `json:"oca"`
	}
}
