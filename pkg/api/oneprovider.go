package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
	"io"
	"net/http"
	"strings"
)

type OneProvider struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewService(endpoint, apiKey, clientKey string) (*OneProvider, error) {
	// TODO: add validation of the service creation
	return &OneProvider{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (s *OneProvider) GetTemplateByName(ctx context.Context, name string) (*VMTemplateResponse, error) {
	var response ListVMTemplatesResponse
	err := s.makeAPICall(ctx, http.MethodGet, "/vm/templates/", nil, &response)
	if err != nil {
		return nil, err
	}
	tpl, found := common.FindElement(response.Templates, func(t VMTemplateResponse) bool {
		return strings.EqualFold(t.Name, name)
	})
	if !found {
		return nil, fmt.Errorf("template not found for name %s", name)
	}
	return &tpl, nil
}

func (s *OneProvider) GetLocationByCity(ctx context.Context, city string) (*VMLocationResponse, error) {
	var response ListVMLocationsResponse
	err := s.makeAPICall(ctx, http.MethodGet, "/vm/locations", nil, &response)
	if err != nil {
		return nil, err
	}

	findCityFn := func(l VMLocationResponse) bool { return l.City == city }

	for _, regions := range response.Response {
		location, found := common.FindElement(regions, findCityFn)
		if found {
			return &location, nil
		}
	}
	return nil, fmt.Errorf("location not found for city %s", city)
}

func (s *OneProvider) GetVMInstanceByID(ctx context.Context, id string) (*VMInstanceReadResponse, error) {
	var response VMInstanceReadResponse
	err := s.makeAPICall(ctx, http.MethodGet, fmt.Sprintf("/vm/info/%s", id), nil, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *OneProvider) CreateVMInstance(ctx context.Context, req *VMInstanceCreateRequest) (*VMInstanceCreateResponse, error) {
	var response VMInstanceCreateResponse

	err := s.makeAPICall(ctx, http.MethodPost, "/vm/create", strings.NewReader(req.UrlValues().Encode()), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *OneProvider) UpdateVMInstance(ctx context.Context, req *VMInstanceUpdateRequest) (*VMInstanceUpdateResponse, error) {
	var response VMInstanceUpdateResponse

	if strings.TrimSpace(req.Hostname) == "" {
		return nil, fmt.Errorf("hostname cannot be empty")
	}

	err := s.makeAPICall(ctx, http.MethodPost, "/vm/hostname", strings.NewReader(req.HostnameUrlValues().Encode()), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *OneProvider) DestroyVMInstance(ctx context.Context, req *VMInstanceDestroyRequest) (*VMInstanceDestroyResponse, error) {
	var response VMInstanceDestroyResponse

	err := s.makeAPICall(ctx, http.MethodPost, "/vm/destroy", strings.NewReader(req.UrlValues().Encode()), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

func (s *OneProvider) makeAPICall(ctx context.Context, method, endpoint string, body io.Reader, result interface{}) error {
	requestURL := fmt.Sprintf("%s%s", s.endpoint, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "OneApi/1.0")
	req.Header.Add("Api-Key", s.apiKey)
	req.Header.Add("Client-Key", s.clientKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// We cannot directly check the response status code
	// because OneProvider API is always sending 200, hiding errors
	// in the response body...
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}

	return nil
}
