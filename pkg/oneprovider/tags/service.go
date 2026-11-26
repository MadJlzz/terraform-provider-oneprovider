package tags

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/client"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/projects"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/oneprovider/vm"
)

type Service struct {
	client   *client.Client
	projects *projects.Service
}

func NewService(c *client.Client, projectService *projects.Service) *Service {
	return &Service{client: c, projects: projectService}
}

func (s *Service) listTags(ctx context.Context, project string) ([]TagReadResponse, error) {
	if project == "" {
		p, err := s.projects.GetDefault(ctx)
		if err != nil {
			return nil, err
		}
		project = p.Uuid
	}
	var response TagListResponse
	err := s.client.MakeAPICall(ctx, http.MethodGet, "/tag/listing/"+url.PathEscape(project), nil, &response)
	if err != nil {
		return nil, fmt.Errorf("tags: list tags in project %s: %w", project, err)
	}
	return response.Tags, nil
}

var ErrNotFound = errors.New("tags: tag not found")

func (s *Service) GetByID(ctx context.Context, project, id string) (*TagReadResponse, error) {
	tags, err := s.listTags(ctx, project)
	if err != nil {
		return nil, err
	}
	if tags == nil {
		return nil, fmt.Errorf("tags: expected a list of tags for project %s", project)
	}
	tag, found := common.FindElement(tags, func(t TagReadResponse) bool { return t.Id.String() == id })
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
	}
	return &tag, nil
}

func (s *Service) Create(ctx context.Context, project, label, description string) (*TagReadResponse, error) {
	form := url.Values{"project_uuid": {project}, "label": {label}, "description": {description}}
	var response TagCreateResponse
	err := s.client.MakeAPICall(ctx, http.MethodPost, "/tag/add", strings.NewReader(form.Encode()), &response)
	if err != nil {
		return nil, fmt.Errorf("tags: create tag: %w", err)
	}
	if response.Tag.Id == "" {
		return nil, fmt.Errorf("tags: create returned an empty tag ID; check project %s for tag %q and import it before retrying", project, label)
	}
	return &response.Tag, nil
}

func (s *Service) Update(ctx context.Context, id, label, description string) error {
	form := url.Values{"tag_id": {id}, "label": {label}, "description": {description}}
	if err := s.client.MakeAPICall(ctx, http.MethodPost, "/tag/edit", strings.NewReader(form.Encode()), nil); err != nil {
		return fmt.Errorf("tags: update tag %s: %w", id, err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	form := url.Values{"tag_id": {id}}
	if err := s.client.MakeAPICall(ctx, http.MethodPost, "/tag/remove", strings.NewReader(form.Encode()), nil); err != nil {
		return fmt.Errorf("tags: delete tag %s: %w", id, err)
	}
	return nil
}

func (s *Service) GetByName(ctx context.Context, project, name string) (*TagReadResponse, error) {
	tags, err := s.listTags(ctx, project)
	if err != nil {
		return nil, err
	}
	findTagFn := func(t TagReadResponse) bool { return t.Label == name }
	tag, found := common.FindElement(tags, findTagFn)
	if !found {
		return nil, fmt.Errorf("tags: tag not found for name %s", name)
	}
	if tag.Id == "" {
		return nil, fmt.Errorf("tags: tag %q has an empty ID", name)
	}
	return &tag, nil
}

func (s *Service) GetInstanceTagIDs(ctx context.Context, project, instanceID string) ([]string, error) {
	tags, err := s.listTags(ctx, project)
	if err != nil {
		return nil, err
	}
	ids := []string{}
	for _, tag := range tags {
		if slices.Contains(tag.Assignees, vm.APIID(instanceID)) {
			if tag.Id == "" {
				return nil, fmt.Errorf("tags: assigned tag has an empty ID")
			}
			ids = append(ids, tag.Id.String())
		}
	}
	return ids, nil
}

func (s *Service) SetInstanceTags(ctx context.Context, project, instanceID string, desired []string) error {
	tags, err := s.listTags(ctx, project)
	if err != nil {
		return err
	}
	// Validate all desired tags before changing any assignments.
	for _, id := range desired {
		_, found := common.FindElement(tags, func(t TagReadResponse) bool { return t.Id.String() == id })
		if id == "" || !found {
			return fmt.Errorf("tags: tag %q not found in the VM's project", id)
		}
	}
	for _, tag := range tags {
		wanted := slices.Contains(desired, tag.Id.String())
		if wanted == slices.Contains(tag.Assignees, vm.APIID(instanceID)) {
			continue
		}
		action := "unassign"
		if wanted {
			action = "assign"
		}
		form := url.Values{"tag_id": {tag.Id.String()}, "service_id": {instanceID}}
		err := s.client.MakeAPICall(ctx, http.MethodPost, "/tag/"+action+"/", strings.NewReader(form.Encode()), nil)
		if err != nil {
			return fmt.Errorf("tags: %s tag %s for instance %s: %w", action, tag.Id, instanceID, err)
		}
	}
	return nil
}
