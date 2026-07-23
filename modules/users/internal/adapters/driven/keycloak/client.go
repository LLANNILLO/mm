package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/llannillo/mm/modules/users/internal/domain"
)

type Config struct {
	AdminURL                 string `mapstructure:"admin_url"`
	TokenURL                 string `mapstructure:"token_url"`
	ConfidentialClientID     string `mapstructure:"confidential_client_id"`
	ConfidentialClientSecret string `mapstructure:"confidential_client_secret"`
}

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, http: &http.Client{}}
}

func (c *Client) RegisterUser(ctx context.Context, email, password, firstName, lastName string) (string, error) {
	token, err := c.getAdminToken(ctx)
	if err != nil {
		return "", fmt.Errorf("get admin token: %w", err)
	}

	body := userRepresentation{
		Username:      email,
		Email:         email,
		FirstName:     firstName,
		LastName:      lastName,
		EmailVerified: true,
		Enabled:       true,
		Credentials: []credentialRepresentation{
			{Type: "password", Value: password, Temporary: false},
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal user representation: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.AdminURL+"users", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("create user: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	if resp.StatusCode == http.StatusConflict {
		return "", domain.ErrEmailAlreadyTaken
	}
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status from Keycloak: %d", resp.StatusCode)
	}

	return extractIdentityID(resp)
}

func (c *Client) getAdminToken(ctx context.Context) (string, error) {
	form := url.Values{
		"client_id":     {c.cfg.ConfidentialClientID},
		"client_secret": {c.cfg.ConfidentialClientSecret},
		"grant_type":    {"client_credentials"},
		"scope":         {"openid"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	return result.AccessToken, nil
}

func extractIdentityID(resp *http.Response) (string, error) {
	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("location header missing from Keycloak response")
	}
	const marker = "users/"
	idx := strings.LastIndex(location, marker)
	if idx < 0 {
		return "", fmt.Errorf("cannot parse identity id from location: %s", location)
	}
	return location[idx+len(marker):], nil
}

type userRepresentation struct {
	Username      string                     `json:"username"`
	Email         string                     `json:"email"`
	FirstName     string                     `json:"firstName"`
	LastName      string                     `json:"lastName"`
	EmailVerified bool                       `json:"emailVerified"`
	Enabled       bool                       `json:"enabled"`
	Credentials   []credentialRepresentation `json:"credentials"`
}

type credentialRepresentation struct {
	Type      string `json:"type"`
	Value     string `json:"value"`
	Temporary bool   `json:"temporary"`
}
