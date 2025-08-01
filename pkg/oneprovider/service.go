package oneprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type API interface {
	ListTemplates(ctx context.Context) (*ListVMTemplatesResponse, error)
	GetLocationByCity(ctx context.Context, city string) (*VMLocationResponse, error)
}

type service struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewService(endpoint, apiKey, clientKey string) (API, error) {
	// TODO: add validation of the service creation
	return &service{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (s *service) makeAPICall(ctx context.Context, method, endpoint string, body io.Reader, result interface{}) error {
	requestURL := fmt.Sprintf("%s%s", s.endpoint, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
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

func (s *service) ListTemplates(ctx context.Context) (*ListVMTemplatesResponse, error) {
	var response ListVMTemplatesResponse
	err := s.makeAPICall(ctx, http.MethodGet, "/vm/templates", nil, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func findLocation(locations map[string][]VMLocationResponse, filterFunc func(response VMLocationResponse) bool) *VMLocationResponse {
	for _, locationList := range locations {
		for _, location := range locationList {
			if filterFunc(location) {
				return &location
			}
		}
	}
	return nil
}

func (s *service) GetLocationByCity(ctx context.Context, city string) (*VMLocationResponse, error) {
	var response ListVMLocationsResponse
	err := s.makeAPICall(ctx, http.MethodGet, "/vm/locations", nil, &response)
	if err != nil {
		return nil, err
	}

	l := findLocation(response.Response, func(l VMLocationResponse) bool {
		return l.City == city
	})
	if l == nil {
		return nil, fmt.Errorf("location not found for city %s", city)
	}
	return l, nil
}
