package ssh

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"net/http"
	"net/url"
	"strings"
)

type Service struct {
	client *client.Client
}

func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

func (s *Service) Create(ctx context.Context, req *CreateSSHKeyRequest) (*CreateSSHKeyResponse, error) {
	var resp CreateSSHKeyResponse

	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/sshkey/new", strings.NewReader(req.UrlValues().Encode()), &resp)
	if err != nil {
		return nil, fmt.Errorf("ssh: create vm sshkey failed: %w", err)
	}

	return &resp, nil
}

func (s *Service) Destroy(ctx context.Context, uuid string) error {
	data := url.Values{"ssh_key": {uuid}}

	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/sshkey/delete", strings.NewReader(data.Encode()), nil)
	if err != nil {
		return fmt.Errorf("ssh: destroy ssh key failed: %w", err)
	}

	return nil
}
