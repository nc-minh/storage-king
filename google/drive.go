package google

import (
	"mime/multipart"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

type Drive interface {
	UpLoad(file multipart.File, fileName string) (*drive.File, error)
}

type DriveUpload struct {
	accessToken string
	ctx         *gin.Context
}

func NewDriveUpload(accessToken string, ctx *gin.Context) Drive {
	return &DriveUpload{
		accessToken: accessToken,
		ctx:         ctx,
	}
}

func (d *DriveUpload) UpLoad(file multipart.File, fileName string) (*drive.File, error) {
	token := &oauth2.Token{
		AccessToken: d.accessToken,
	}
	client := oauth2.NewClient(d.ctx, nil)
	client.Transport = &oauth2.Transport{
		Base:   client.Transport,
		Source: oauth2.ReuseTokenSource(nil, oauth2.StaticTokenSource(token)),
	}

	driveFile := &drive.File{
		Name:    filepath.Base(fileName),
		Parents: []string{"root"},
	}

	srv, err := drive.NewService(d.ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Err(err).Msg("unable to retrieve Drive client")
	}

	res, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		log.Err(err).Msg("unable to create file")
		return nil, err
	}

	// Set the permission for the new file
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = srv.Permissions.Create(res.Id, permission).Do()
	if err != nil {
		log.Err(err).Msg("unable to create permission")
	}

	return res, nil
}
