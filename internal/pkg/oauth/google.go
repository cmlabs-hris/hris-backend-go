package oauth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleService interface {
	// GenerateState generates a random state string for OAuth2 flows.
	GenerateState(userAgent string) string
	// RedirectURL generates the OAuth2 redirect URL with a state.
	RedirectURL(state string) string
	// VerifyToken exchanges the code for an OAuth2 token.
	VerifyToken(ctx context.Context, code string) (*oauth2.Token, error)
	// VerifyUser fetches and verifies the Google user information.
	VerifyUser(ctx context.Context, token *oauth2.Token) (GoogleInformation, error)
}

type GoogleServiceImpl struct {
	config *oauth2.Config
}

func NewGoogleService(clientID string, clientSecret string, redirectURL string, scopes []string) GoogleService {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       scopes,
		Endpoint:     google.Endpoint,
	}
	return &GoogleServiceImpl{config: config}
}

type GoogleInformation struct {
	GoogleID      string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
}

// GenerateState generates a random state string for OAuth2 flows.
func (g *GoogleServiceImpl) GenerateState(userAgent string) string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	state := fmt.Sprintf("%s.%s", base64.URLEncoding.EncodeToString(b), userAgent)
	return base64.URLEncoding.EncodeToString([]byte(state))
}

func (g *GoogleServiceImpl) RedirectURL(state string) string {
	return g.config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func (g *GoogleServiceImpl) VerifyToken(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := g.config.Exchange(ctx, code)
	if err != nil {
		return &oauth2.Token{}, err
	}
	return token, nil
}

func (g *GoogleServiceImpl) VerifyUser(ctx context.Context, token *oauth2.Token) (GoogleInformation, error) {
	var req GoogleInformation

	client := g.config.Client(ctx, token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return GoogleInformation{}, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&req); err != nil {
		return GoogleInformation{}, err
	}

	return req, nil
}
