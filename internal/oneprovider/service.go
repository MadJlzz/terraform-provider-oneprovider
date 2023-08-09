package oneprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type API interface {
	ListTemplates(ctx context.Context) (*ListTemplatesResponse, error)
}

type service struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewService(endpoint, apiKey, clientKey string) (API, error) {
	return &service{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (s *service) ListTemplates(ctx context.Context) (*ListTemplatesResponse, error) {
	url := fmt.Sprintf("%s/vm/templates", s.endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	s.addHeaders(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var ltr ListTemplatesResponse
	err = json.NewDecoder(resp.Body).Decode(&ltr)
	if err != nil {
		return nil, err
	}

	return &ltr, nil
}

func (s *service) addHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("User-Agent", "OneApi/1.0")
	req.Header.Add("Api-Key", s.apiKey)
	req.Header.Add("Client-Key", s.clientKey)
}
