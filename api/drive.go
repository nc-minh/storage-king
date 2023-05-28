package api

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func (server *Server) uploadToDrive(driveFile *drive.File, file multipart.File, accessToken string) (*drive.Service, *drive.File, error) {
	log.Info().Msg("Uploading file to Google Drive...")
	log.Warn().Msgf("accessToken %s", accessToken)
	fmt.Println("driveFile", driveFile)
	fmt.Println("file", file)

	client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	}))

	srv, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		log.Err(err).Msg("unable to retrieve Drive client")
		return nil, nil, err
	}

	driveFile.MimeType = "image/jpeg"
	res, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		log.Err(err).Msg("unable to create file")
		return nil, nil, err
	}

	log.Info().Msg(fmt.Sprintf("File '%s' successfully uploaded.", res.Name))
	return srv, res, nil
}

func (server *Server) upload(ctx *gin.Context) {

	id := ctx.PostForm("id")
	email := ctx.PostForm("email")

	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err).Msg("unable to parse id")
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		log.Err(err).Msg("unable to get file")
		return
	}
	defer file.Close()

	tempFile, err := ioutil.TempFile("", "temp-file-*.tmp")
	if err != nil {
		log.Err(err).Msg("unable to create temporary file")
		return
	}
	defer os.Remove(tempFile.Name()) // Xóa file tạm thời sau khi hoàn thành

	// Sao chép nội dung của file vào file tạm thời
	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Err(err).Msg("unable to copy file to temporary file")
		return
	}

	// Đặt lại con trỏ của file tạm thời về đầu
	tempFile.Seek(0, 0)

	filename := header.Filename
	driveFile := &drive.File{
		Name:    filepath.Base(filename),
		Parents: []string{"root"},
	}

	arg := db.GetStorageParams{
		ID:    num,
		Email: email,
	}

	storage, err := server.store.GetStorage(ctx, arg)
	if err != nil {
		log.Err(err).Msg("unable to get storage")
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	log.Info().Msgf("essToken %s", storage.AccessToken)

	var srv *drive.Service
	var res *drive.File

	srv, res, err = server.uploadToDrive(driveFile, tempFile, storage.AccessToken)
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 401 {
			refreshTokenResponse, err := server.google.RefreshToken(storage.RefreshToken, server.config.ClientID, server.config.ClientSecret, arg.ID)
			if err != nil {
				log.Err(err).Msg("unable to refresh token")
				ctx.JSON(http.StatusInternalServerError, errorResponse(err))
				return
			}

			log.Info().Msgf("refresh token successfully %s", refreshTokenResponse.AccessToken)

			srv, res, err = server.uploadToDrive(driveFile, tempFile, refreshTokenResponse.AccessToken)
			if err != nil {
				log.Err(err).Msg("unable to upload to drive")
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"message": "unauthorized",
				})
				return
			}
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

	fmt.Printf("File '%s' uploaded successfully to parent folder '%s'.\n", res.Name, res.Id)

	ctx.JSON(200, gin.H{
		"message": "uploaded successfully",
		"res":     res,
	})
}
