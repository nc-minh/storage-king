package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type GoogleAuthService interface {
	CreateAuthURL() string
	CodeAuthentication(code string) (*oauth2.Token, error)
	GetEmailFromAccessToken(accessToken string) (string, error)
	RefreshToken(refreshToken string, clientId string, clientSecret string) error
}

type GoogleAuth struct {
	oauth2Config *oauth2.Config
	store        db.Store
}

func NewDrive(oauth2Config *oauth2.Config, store db.Store) GoogleAuthService {
	return &GoogleAuth{
		oauth2Config: oauth2Config,
		store:        store,
	}
}

func (g *GoogleAuth) CreateAuthURL() string {
	authURL := g.oauth2Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Info().Msgf("create authURL success")
	return authURL
}

func (d *GoogleAuth) CodeAuthentication(code string) (*oauth2.Token, error) {
	tok, err := d.oauth2Config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatal().Msg("unable to retrieve token from web")
		return nil, err
	}
	return tok, nil
}

func (g *GoogleAuth) GetEmailFromAccessToken(accessToken string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		log.Error().Msgf("failed to create userinfo request: %v", err)
		return "", err
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("failed to get userinfo: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("failed to get userinfo: %s", resp.Status)
		return "", errors.New("failed to get userinfo")
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		log.Error().Msgf("failed to decode userinfo: %v", err)
		return "", err
	}

	return userInfo.Email, nil
}

type createFreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

func (g *GoogleAuth) RefreshToken(refreshToken string, clientId string, clientSecret string) error {
	client := &http.Client{}

	bodyData := createFreshTokenRequest{
		GrantType:    "refresh_token",
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
	body, err := json.Marshal(bodyData)
	if err != nil {
		log.Error().Msgf("failed to marshal request body: %v", err)
		return err
	}

	req, err := http.NewRequest("POST", "https://www.googleapis.com/oauth2/v4/token", bytes.NewReader(body))
	if err != nil {
		log.Error().Msgf("faile to refresh token: %v", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("failed to get data: %v", err)
		return err
	}
	defer resp.Body.Close()
	return nil
}
