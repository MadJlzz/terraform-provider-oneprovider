package oneprovider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

var (
	ErrEndpointMalformed  = errors.New("endpoint is a malformed URI")
	ErrApiKeyMalformed    = errors.New("api key should not be empty")
	ErrClientKeyMalformed = errors.New("client key should not be empty")
)

type API interface {
	ListTemplates(ctx context.Context) (*ListVMTemplatesResponse, error)
	ListLocations(ctx context.Context) (*ListLocationsResponse, error)
}

type service struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewService(endpoint, apiKey, clientKey string) (API, error) {
	_, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, ErrEndpointMalformed
	}

	if strings.TrimSpace(apiKey) == "" {
		return nil, ErrApiKeyMalformed
	}

	if strings.TrimSpace(clientKey) == "" {
		return nil, ErrClientKeyMalformed
	}

	return &service{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (s *service) ListTemplates(ctx context.Context) (*ListVMTemplatesResponse, error) {
	uri := fmt.Sprintf("%s/vm/templates", s.endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ltr ListVMTemplatesResponse
	err = json.NewDecoder(resp.Body).Decode(&ltr)
	if err != nil {
		return nil, err
	}

	return &ltr, nil
}

func (s *service) ListLocations(ctx context.Context) (*ListLocationsResponse, error) {
	uri := fmt.Sprintf("%s/vm/locations", s.endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var llr ListLocationsResponse
	err = json.NewDecoder(resp.Body).Decode(&llr)
	if err != nil {
		return nil, err
	}

	return &llr, nil
}

func (s *service) addHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "OneApi/1.0")
	req.Header.Add("Api-Key", s.apiKey)
	req.Header.Add("Client-Key", s.clientKey)
}
