package oneprovider

import (
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/ssh"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
)

type OneProvider struct {
	VM  *vm.Service
	SSH *ssh.Service
}

func NewService(endpoint, apiKey, clientKey string) (*OneProvider, error) {
	// TODO: add validation of the service creation
	c, err := client.NewClient(endpoint, apiKey, clientKey)
	if err != nil {
		return nil, err
	}

	return &OneProvider{
		VM:  vm.NewService(c),
		SSH: ssh.NewService(c),
	}, nil
}
