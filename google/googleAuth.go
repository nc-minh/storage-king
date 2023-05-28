package google

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type GoogleAuthService interface {
	CreateAuthURL() string
	CodeAuthentication(code string) (*oauth2.Token, error)
	GetEmailFromAccessToken(accessToken string) (string, error)
	RefreshToken(refreshToken string, clientId string, clientSecret string, storageId int64) (RefreshTokenReponse, error)
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

type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type"`
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenReponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IdToken     string `json:"id_token"`
}

func (g *GoogleAuth) RefreshToken(refreshToken string, clientId string, clientSecret string, storageId int64) (RefreshTokenReponse, error) {
	client := &http.Client{}

	bodyData := RefreshTokenRequest{
		GrantType:    "refresh_token",
		ClientId:     clientId,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
	body, err := json.Marshal(bodyData)
	if err != nil {
		log.Error().Msgf("failed to marshal request body: %v", err)
		return RefreshTokenReponse{}, err
	}

	req, err := http.NewRequest("POST", "https://www.googleapis.com/oauth2/v4/token", bytes.NewReader(body))
	if err != nil {
		log.Error().Msgf("faile to refresh token: %v", err)
		return RefreshTokenReponse{}, err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Error().Msgf("failed to get data: %v", err)
		return RefreshTokenReponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Error().Msgf("request failed with status code: %d", resp.StatusCode)
		return RefreshTokenReponse{}, errors.New("request failed")
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Msgf("failed to read response body: %v", err)
		return RefreshTokenReponse{}, err
	}

	var refreshTokenResponse RefreshTokenReponse
	err = json.Unmarshal(body, &refreshTokenResponse)
	if err != nil {
		log.Error().Msgf("failed to unmarshal response body: %v", err)
		return RefreshTokenReponse{}, err
	}

	_, err = g.store.UpdateStorage(context.Background(), db.UpdateStorageParams{
		ID: storageId,
		AccessToken: sql.NullString{
			String: refreshTokenResponse.AccessToken,
			Valid:  true,
		},
		IsRefreshTokenExpired: sql.NullBool{
			Bool:  false,
			Valid: true,
		},
	})

	if err != nil {
		log.Error().Msgf("failed to update storage: %v", err)
		return RefreshTokenReponse{}, err
	}

	return refreshTokenResponse, nil
}
