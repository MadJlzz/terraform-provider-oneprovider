package ssh

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"net/http"
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
