package oneprovider

import (
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/projects"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/ssh"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/tags"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
)

type Service struct {
	VM       *vm.Service
	SSH      *ssh.Service
	Tags     *tags.Service
	Projects *projects.Service
}

func NewService(endpoint, apiKey, clientKey string) (*Service, error) {
	c, err := client.NewClient(endpoint, apiKey, clientKey)
	if err != nil {
		return nil, err
	}
	projectService := projects.NewService(c)
	return &Service{
		VM:       vm.NewService(c),
		SSH:      ssh.NewService(c),
		Tags:     tags.NewService(c, projectService),
		Projects: projectService,
	}, nil
}
