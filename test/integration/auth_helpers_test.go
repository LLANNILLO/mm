package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// getAccessToken performs a Resource Owner Password Credentials grant against
// the public client, mirroring BaseIntegrationTest.GetAccessTokenAsync in the
// C# reference.
func getAccessToken(email, password string) (string, error) {
	form := url.Values{
		"client_id":  {keycloakPublicClientID},
		"grant_type": {"password"},
		"scope":      {"openid"},
		"username":   {email},
		"password":   {password},
	}

	resp, err := httpClient.Post(keycloakTokenURL, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned %d", resp.StatusCode)
	}

	var body struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}
	return body.AccessToken, nil
}
