package api

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

var driveClient = make(map[string]*drive.Service)

func (server *Server) getDriveClient(ctx *gin.Context, accessToken string) (*drive.Service, error) {
	id := ctx.PostForm("id")

	if driveClient[id] != nil {
		return driveClient[id], nil
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))
	log.Info().Msg("Drive client created")

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Err(err).Msg("unable to retrieve Drive client")
		return nil, err
	}

	driveClient[id] = srv
	return driveClient[id], nil
}

func (server *Server) uploadToDrive(ctx *gin.Context, driveFile *drive.File, file multipart.File, accessToken string) (*drive.Service, *drive.File, error) {
	srv, err := server.getDriveClient(ctx, accessToken)
	if err != nil {
		return nil, nil, err
	}

	res, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		log.Err(err).Msg("unable to create file with Drive client")
		return nil, nil, err
	}

	log.Info().Msg(fmt.Sprintf("File '%s' successfully uploaded.", res.Name))
	return srv, res, nil
}

func (server *Server) upload(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Err(err).Msg("unable to get file")
		return
	}
	defer file.Close()

	filename := header.Filename
	driveFile := &drive.File{
		Name:    filepath.Base(filename),
		Parents: []string{"root"},
	}

	var srv *drive.Service
	var res *drive.File

	accessToken := ctx.GetString("access_token")

	srv, res, err = server.uploadToDrive(ctx, driveFile, file, accessToken)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 401 {
			log.Err(err).Msg("unable to upload to drive due to invalid token")
			return
		} else {
			log.Err(err).Msg("unable to upload to drive")
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			return
		}
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

	ctx.JSON(200, gin.H{
		"message": "uploaded successfully",
		"res":     res,
	})
}
