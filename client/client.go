package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	Address    string
	Token      string
	HTTPClient *http.Client
}

func NewClient(address, token string) *Client {
	return &Client{
		Address: address,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) sendRequest(ctx context.Context, method, path string, bodyVal interface{}, respVal interface{}) error {
	var bodyReader *bytes.Reader
	if bodyVal != nil {
		data, err := json.Marshal(bodyVal)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(data)
	}

	url := fmt.Sprintf("%s%s", c.Address, path)
	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	if bodyVal != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			if msg, exists := errResp["error"]; exists {
				return fmt.Errorf("API error (status %d): %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("API error (status %d)", resp.StatusCode)
	}

	if respVal != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(respVal)
	}

	return nil
}

type TeamConfig struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type CertConfig struct {
	ID          string   `json:"id,omitempty"`
	Primary     string   `json:"primary"`
	Sans        []string `json:"sans,omitempty"`
	TeamID      string   `json:"team_id"`
	Description string   `json:"description,omitempty"`
	DNSProvider string   `json:"dns_provider,omitempty"`
}

type APIKeyConfig struct {
	ID                  string   `json:"id,omitempty"`
	Token               string   `json:"token,omitempty"`
	CleartextToken      string   `json:"cleartext_token,omitempty"`
	Description         string   `json:"description,omitempty"`
	AllowedCertificates []string `json:"allowed_certificates,omitempty"`
	AllowedTeams        []string `json:"allowed_teams,omitempty"`
	Admin               bool     `json:"admin"`
}

type CertificateData struct {
	ID           string   `json:"id"`
	Domain       string   `json:"domain"`
	Sans         []string `json:"sans"`
	Issued       bool     `json:"issued"`
	Certificate  string   `json:"certificate,omitempty"`
	PrivateKey   string   `json:"private_key,omitempty"`
	CertFilename string   `json:"cert_filename,omitempty"`
	KeyFilename  string   `json:"key_filename,omitempty"`
}

// Certificate Config CRUD
func (c *Client) GetCertificates(ctx context.Context) ([]CertConfig, error) {
	var resp []CertConfig
	err := c.sendRequest(ctx, "GET", "/api/v1/config/certificates", nil, &resp)
	return resp, err
}

func (c *Client) CreateCertificate(ctx context.Context, cert CertConfig) (CertConfig, error) {
	var resp CertConfig
	err := c.sendRequest(ctx, "POST", "/api/v1/config/certificates", cert, &resp)
	return resp, err
}

func (c *Client) UpdateCertificate(ctx context.Context, id string, cert CertConfig) error {
	path := fmt.Sprintf("/api/v1/config/certificates/%s", id)
	return c.sendRequest(ctx, "PUT", path, cert, nil)
}

func (c *Client) DeleteCertificate(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/config/certificates/%s", id)
	return c.sendRequest(ctx, "DELETE", path, nil, nil)
}

// API Key Config CRUD
func (c *Client) GetAPIKeys(ctx context.Context) ([]APIKeyConfig, error) {
	var resp []APIKeyConfig
	err := c.sendRequest(ctx, "GET", "/api/v1/config/api_keys", nil, &resp)
	return resp, err
}

func (c *Client) CreateAPIKey(ctx context.Context, key APIKeyConfig) (APIKeyConfig, error) {
	var resp APIKeyConfig
	err := c.sendRequest(ctx, "POST", "/api/v1/config/api_keys", key, &resp)
	return resp, err
}

func (c *Client) UpdateAPIKey(ctx context.Context, key APIKeyConfig) error {
	path := fmt.Sprintf("/api/v1/config/api_keys/%s", key.ID)
	return c.sendRequest(ctx, "PUT", path, key, nil)
}

func (c *Client) DeleteAPIKey(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/config/api_keys/%s", id)
	return c.sendRequest(ctx, "DELETE", path, nil, nil)
}

func (c *Client) GetCertificateData(ctx context.Context) ([]CertificateData, error) {
	var resp []CertificateData
	err := c.sendRequest(ctx, "GET", "/api/v1/certificates", nil, &resp)
	return resp, err
}

// Team Config CRUD
func (c *Client) GetTeams(ctx context.Context) ([]TeamConfig, error) {
	var resp []TeamConfig
	err := c.sendRequest(ctx, "GET", "/api/v1/config/teams", nil, &resp)
	return resp, err
}

func (c *Client) CreateTeam(ctx context.Context, team TeamConfig) (TeamConfig, error) {
	var resp TeamConfig
	err := c.sendRequest(ctx, "POST", "/api/v1/config/teams", team, &resp)
	return resp, err
}

func (c *Client) UpdateTeam(ctx context.Context, id string, team TeamConfig) error {
	path := fmt.Sprintf("/api/v1/config/teams/%s", id)
	return c.sendRequest(ctx, "PUT", path, team, nil)
}

func (c *Client) DeleteTeam(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/v1/config/teams/%s", id)
	return c.sendRequest(ctx, "DELETE", path, nil, nil)
}

