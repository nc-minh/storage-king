package api

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

var (
	googleOauthConfig *oauth2.Config
	oauthStateString  = "thisshouldberandom"
)

func (server *Server) createDrive(ctx *gin.Context) {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     server.config.ClientId,
		ClientSecret: server.config.ClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/drive",
		},
		Endpoint: google.Endpoint,
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving the file: %s", err.Error()))
		return
	}
	defer file.Close()
	filename := header.Filename
	client := getClient()
	driveService, err := drive.New(client)
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error getting Drive client: %v", err))
		return
	}
	f := &drive.File{Name: filename}
	createdFile, err := driveService.Files.Create(f).Media(file).Do()
	if err != nil {
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error uploading file: %v", err))
		return
	}

	ctx.String(http.StatusOK, fmt.Sprintf("File '%s' created with ID '%s'.", createdFile.Name, createdFile.Id))
}

func getClient() *http.Client {
	ctx := context.Background()
	token := getTokenFromWeb(ctx)
	return googleOauthConfig.Client(ctx, token)
}

func getTokenFromWeb(ctx context.Context) *oauth2.Token {
	authURL := googleOauthConfig.AuthCodeURL(oauthStateString)
	fmt.Printf("Go to the following link in your browser: %v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := googleOauthConfig.Exchange(ctx, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return token
}

func getAccessToken() (*oauth2.Token, error) {
	clientID := "YOUR_CLIENT_ID"
	clientSecret := "YOUR_CLIENT_SECRET"
	accessTokenURL := "https://oauth2.googleapis.com/token"

	ctx := context.Background()
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  "urn:ietf:wg",
		Scopes:       []string{"https://www.googleapis.com/auth/drive"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: accessTokenURL,
		},
	}

	if err != nil {
		return nil, err
	}

	return token, nil
}
