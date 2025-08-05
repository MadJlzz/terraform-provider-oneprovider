package ssh

import (
	"context"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
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

func (s *Service) GetByID(ctx context.Context, id string) (*SshKeyReadResponse, error) {
	var resp SshKeyListResponse

	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/sshkeys/list", nil, &resp)
	if err != nil {
		return nil, fmt.Errorf("ssh: list ssh keys failed: %w", err)
	}

	key, found := common.FindElement(resp.Response.SshKeys, func(k SshKeyReadResponse) bool {
		return id == k.Uuid
	})

	if !found {
		return nil, fmt.Errorf("ssh: key not found for id %s", id)
	}

	return &key, nil
}

func (s *Service) Create(ctx context.Context, req *SshKeyCreateRequest) (*SshKeyCreateResponse, error) {
	var resp SshKeyCreateResponse

	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/sshkey/new", strings.NewReader(req.UrlValues().Encode()), &resp)
	if err != nil {
		return nil, fmt.Errorf("ssh: create vm sshkey failed: %w", err)
	}

	return &resp, nil
}

type SshKeyUpdateRequest struct {
	Uuid      string `json:"uuid"`
	Name      string `json:"name"`
	PublicKey string `json:"value"`
}

func (v *SshKeyUpdateRequest) UrlValues() url.Values {
	return url.Values{
		"ssh_key":   {v.Uuid},
		"key_name":  {v.Name},
		"key_value": {v.PublicKey},
	}
}

type SshKeyUpdateResponse struct {
	Response struct {
		SshKeys []struct {
			Name      string `json:"name"`
			PublicKey string `json:"value"`
		}
	}
}

func (s *Service) Update(ctx context.Context, req *SshKeyUpdateRequest) (*SshKeyUpdateResponse, error) {
	var resp SshKeyUpdateResponse

	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/sshkey/edit", strings.NewReader(req.UrlValues().Encode()), &resp)
	if err != nil {
		return nil, fmt.Errorf("ssh: update ssh key failed: %w", err)
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
