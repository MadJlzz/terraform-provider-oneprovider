package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MadJlzz/terraform-provider-oneprovider/pkg/common"
)

type Client struct {
	apiKey    string
	clientKey string
	endpoint  string
}

func NewClient(endpoint, apiKey, clientKey string) (*Client, error) {
	endpoint = strings.TrimSpace(endpoint)
	apiKey = strings.TrimSpace(apiKey)
	clientKey = strings.TrimSpace(clientKey)

	if endpoint == "" {
		return nil, fmt.Errorf("client: endpoint cannot be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("client: apiKey cannot be empty")
	}
	if clientKey == "" {
		return nil, fmt.Errorf("client: clientKey cannot be empty")
	}

	return &Client{
		endpoint:  endpoint,
		apiKey:    apiKey,
		clientKey: clientKey,
	}, nil
}

func (c *Client) MakeAPICall(ctx context.Context, method, endpoint string, body io.Reader, result any) error {
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
		return fmt.Errorf("client: api request failed with status: %d", resp.StatusCode)
	}

	// Read the entire response body first
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("client: failed to read response body: %w", err)
	}

	// Check for API errors in the response body
	var errorCheck struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	// Try to decode the error structure.
	unmarshalErr := json.Unmarshal(bodyBytes, &errorCheck)
	if unmarshalErr != nil {
		return fmt.Errorf("client: failed to decode error response body: %w", unmarshalErr)
	}

	// If there is an error, we stop and return it.
	if errorCheck.Error != nil {
		switch errorCheck.Error.Code {
		case 42:
			return common.ErrVmNotFound
		default:
			return fmt.Errorf("client: api internal error %d: %s", errorCheck.Error.Code, errorCheck.Error.Message)
		}
	}

	if result != nil {
		// No API error found, decode into the result interface
		if decodeErr := json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(result); decodeErr != nil {
			return fmt.Errorf("client: failed to decode response: %w", decodeErr)
		}
	}

	return nil
}
