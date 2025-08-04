package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Client struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewClient(endpoint, apiKey, clientKey string) (*Client, error) {
	// TODO: add validation of the client creation
	return &Client{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (c *Client) MakeAPICall(ctx context.Context, method, endpoint string, body io.Reader, result interface{}) error {
	requestURL := fmt.Sprintf("%s%s", c.endpoint, endpoint)

	req, err := http.NewRequestWithContext(ctx, method, requestURL, body)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "OneApi/1.0")
	req.Header.Add("Api-Key", c.apiKey)
	req.Header.Add("Client-Key", c.clientKey)

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
