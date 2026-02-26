package portainer

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	requestTimeout = 15 * time.Second
	maxBodyBytes   = 10 * 1024 * 1024 // 10 MB — generous for any compose file
)

// Stack represents a Portainer stack entry.
type Stack struct {
	ID   int    `json:"Id"`
	Name string `json:"Name"`
}

type stackFileResponse struct {
	StackFileContent string `json:"StackFileContent"`
}

// Client is a minimal Portainer API client.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient returns a Client. It accepts both http:// and https:// URLs.
// For self-hosted Portainer with a self-signed cert, TLS verification is kept
// enabled by default — users should use a valid cert or a local CA.
func NewClient(baseURL, token string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		DisableKeepAlives: true, // single-shot tool — no need to keep connections open
	}

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout:   requestTimeout,
			Transport: transport,
		},
	}
}

// GetStacks returns all stacks visible to the token.
func (c *Client) GetStacks() ([]Stack, error) {
	body, err := c.get("/api/stacks")
	if err != nil {
		return nil, err
	}

	var stacks []Stack
	if err := json.Unmarshal(body, &stacks); err != nil {
		return nil, fmt.Errorf("parse stacks response: %w", err)
	}

	return stacks, nil
}

// GetStackFile returns the compose YAML for a given stack ID.
func (c *Client) GetStackFile(stackID int) ([]byte, error) {
	body, err := c.get(fmt.Sprintf("/api/stacks/%d/file", stackID))
	if err != nil {
		return nil, err
	}

	var resp stackFileResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse stack file response: %w", err)
	}

	if resp.StackFileContent == "" {
		return nil, fmt.Errorf("empty compose content returned by Portainer")
	}

	return []byte(resp.StackFileContent), nil
}

// get performs a GET request against the Portainer API and returns the body.
func (c *Client) get(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set("X-API-Key", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("authentication failed (HTTP %d) — check your token", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP %d from Portainer", resp.StatusCode)
	}

	limited := io.LimitReader(resp.Body, maxBodyBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return body, nil
}
