package projects

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
)

var ErrNotFound = errors.New("projects: project not found")

type Service struct {
	client *client.Client
}

func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

func (s *Service) list(ctx context.Context) ([]ProjectReadResponse, error) {
	var response ProjectsListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/project/list", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("projects: list projects failed: %w", err)
	}
	if response.Response.Projects == nil {
		return nil, fmt.Errorf("projects: expected a project list in the API response")
	}
	return response.Response.Projects, nil
}

func (s *Service) GetDefault(ctx context.Context) (*ProjectReadResponse, error) {
	projects, err := s.list(ctx)
	if err != nil {
		return nil, err
	}
	findDefaultFn := func(p ProjectReadResponse) bool { return p.IsDefault == 1 }
	project, found := common.FindElement(projects, findDefaultFn)
	if !found {
		return nil, fmt.Errorf("projects: default project not found")
	}
	if project.Uuid == "" {
		return nil, fmt.Errorf("projects: default project has an empty UUID")
	}
	return &project, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*ProjectReadResponse, error) {
	projects, err := s.list(ctx)
	if err != nil {
		return nil, err
	}
	findProjectFn := func(p ProjectReadResponse) bool { return p.Name == name }
	project, found := common.FindElement(projects, findProjectFn)
	if !found {
		return nil, fmt.Errorf("%w for name %s", ErrNotFound, name)
	}
	return &project, nil
}

func (s *Service) GetByID(ctx context.Context, uuid string) (*ProjectReadResponse, error) {
	projects, err := s.list(ctx)
	if err != nil {
		return nil, err
	}
	findProjectFn := func(p ProjectReadResponse) bool { return p.Uuid == uuid }
	project, found := common.FindElement(projects, findProjectFn)
	if !found {
		return nil, fmt.Errorf("%w for UUID %s", ErrNotFound, uuid)
	}
	return &project, nil
}

func projectForm(name string, isDefault bool) url.Values {
	flag := "0"
	if isDefault {
		flag = "1"
	}
	return url.Values{"name": {name}, "is_default": {flag}}
}

func (s *Service) Create(ctx context.Context, name string, isDefault bool) (*ProjectReadResponse, error) {
	var response ProjectCreateResponse
	form := projectForm(name, isDefault)
	if err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/project/new", strings.NewReader(form.Encode()), &response); err != nil {
		return nil, fmt.Errorf("projects: create project: %w", err)
	}
	if response.Response.Project.Uuid == "" {
		return nil, fmt.Errorf("projects: create returned an empty project UUID; check for project %q and import it before retrying", name)
	}
	return &response.Response.Project, nil
}

func (s *Service) Update(ctx context.Context, uuid, name string, isDefault bool) error {
	form := projectForm(name, isDefault)
	form.Set("project_uuid", uuid)
	if err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/project/edit", strings.NewReader(form.Encode()), nil); err != nil {
		return fmt.Errorf("projects: update project %s: %w", uuid, err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, uuid string) error {
	form := url.Values{"project_uuid": {uuid}}
	if err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/project/delete", strings.NewReader(form.Encode()), nil); err != nil {
		return fmt.Errorf("projects: delete project %s: %w", uuid, err)
	}
	return nil
}
