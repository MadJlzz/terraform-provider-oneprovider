package vm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
)

type Service struct {
	client *client.Client
}

func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

func (s *Service) GetTemplateByName(ctx context.Context, name string) (*TemplateReadResponse, error) {
	var response TemplatesListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/templates/", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("vm: get template by name failed: %w", err)
	}
	tpl, found := common.FindElement(response.Templates, func(t TemplateReadResponse) bool {
		return strings.EqualFold(t.Name, name)
	})
	if !found {
		return nil, fmt.Errorf("vm: template not found for name %s", name)
	}
	return &tpl, nil
}

func (s *Service) GetLocationByCity(ctx context.Context, city string) (*LocationReadResponse, error) {
	var response LocationsListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/locations", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("vm: get location by city name failed: %w", err)
	}

	findCityFn := func(l LocationReadResponse) bool { return l.City == city }

	for _, regions := range response.Response {
		location, found := common.FindElement(regions, findCityFn)
		if found {
			return &location, nil
		}
	}
	return nil, fmt.Errorf("vm: location not found for city %s", city)
}

func (s *Service) GetInstanceByID(ctx context.Context, id string) (*InstanceReadResponse, error) {
	var response InstanceReadResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, fmt.Sprintf("/vm/info/%s", id), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("vm: get instance by ID failed: %w", err)
	}

	return &response, nil
}

func (s *Service) CreateInstance(ctx context.Context, req *InstanceCreateRequest) (*InstanceCreateResponse, error) {
	var response InstanceCreateResponse

	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/create", strings.NewReader(req.UrlValues().Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("vm: create instance failed: %w", err)
	}

	return &response, nil
}

func (s *Service) UpdateInstanceHostname(ctx context.Context, req *InstanceHostnameUpdateRequest) error {
	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/hostname", strings.NewReader(req.UrlValues().Encode()), nil)
	if err != nil {
		return fmt.Errorf("vm: update instance hostname failed: %w", err)
	}
	return nil
}

func (s *Service) UpdateInstanceProject(ctx context.Context, req *InstanceProjectUpdateRequest) error {
	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/project/assign", strings.NewReader(req.UrlValues().Encode()), nil)
	if err != nil {
		return fmt.Errorf("vm: update instance project failed: %w", err)
	}
	return nil
}

func (s *Service) DestroyInstance(ctx context.Context, req *InstanceDestroyRequest) error {
	err := s.client.MakeAPICall(ctx, http.MethodPost, "/vm/destroy", strings.NewReader(req.UrlValues().Encode()), nil)
	if err != nil {
		return fmt.Errorf("vm: destroy instance failed: %w", err)
	}
	return nil
}

func (s *Service) listProjects(ctx context.Context) ([]ProjectReadResponse, error) {
	var response ProjectsListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/project/list", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("vm: list projects failed: %w", err)
	}
	return response.Response.Projects, nil
}

func (s *Service) GetDefaultProject(ctx context.Context) (*ProjectReadResponse, error) {
	projects, err := s.listProjects(ctx)
	if err != nil {
		return nil, err
	}
	findDefaultFn := func(p ProjectReadResponse) bool { return p.IsDefault == 1 }
	project, found := common.FindElement(projects, findDefaultFn)
	if !found {
		return nil, fmt.Errorf("vm: default project not found")
	}
	if project.Uuid == "" {
		return nil, fmt.Errorf("vm: default project has an empty UUID")
	}
	return &project, nil
}

func (s *Service) GetProjectByName(ctx context.Context, name string) (*ProjectReadResponse, error) {
	projects, err := s.listProjects(ctx)
	if err != nil {
		return nil, err
	}
	findProjectFn := func(p ProjectReadResponse) bool { return p.Name == name }
	project, found := common.FindElement(projects, findProjectFn)
	if found {
		return &project, nil
	}

	return nil, fmt.Errorf("vm: project not found for name %s", name)
}

func (s *Service) GetSizeByName(ctx context.Context, name string) (*SizeReadResponse, error) {
	var response SizesListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/vm/sizes", nil, &response)
	if err != nil {
		return nil, fmt.Errorf("vm: get size by name failed: %w", err)
	}

	findSizeFn := func(s SizeReadResponse) bool { return s.Name == name }
	size, found := common.FindElement(response.Response, findSizeFn)
	if found {
		return &size, nil
	}

	return nil, fmt.Errorf("vm: size not found for name %s", name)
}
