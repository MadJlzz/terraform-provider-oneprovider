package tags

import "github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"

type TagReadResponse struct {
	Id          vm.APIID   `json:"id"`
	Label       string     `json:"label"`
	Description string     `json:"description"`
	ProjectUuid string     `json:"project_uuid"`
	Assignees   []vm.APIID `json:"assignees"`
}

type TagCreateResponse struct {
	Tag TagReadResponse `json:"response"`
}

type TagListResponse struct {
	Tags []TagReadResponse `json:"response"`
}
