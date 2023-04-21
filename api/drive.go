package api

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

func (server *Server) getClient(ctx *gin.Context) *http.Client {
	fmt.Println("getClient is running")
	tokFile := "token.json"
	tok, err := server.getTokenFromFile(ctx, tokFile)
	if err != nil {
		tok = server.getTokenFromWeb(ctx)
		server.saveToken(ctx, tokFile, tok)
	}
	return server.oauth2Config.Client(ctx, tok)
}

func (server *Server) upload(ctx *gin.Context) {

	client := server.getClient(ctx)
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.String(http.StatusBadRequest, fmt.Sprintf("Error retrieving the file: %s", err.Error()))
		return
	}
	defer file.Close()

	filename := header.Filename

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)
	}

	// Create a Drive file.
	driveFile := &drive.File{
		Name:    filepath.Base(filename),
		Parents: []string{"root"},
	}

	// Upload the file to Drive.
	res, err := srv.Files.Create(driveFile).Media(file).Do()
	if err != nil {
		fmt.Println(err)
		ctx.String(http.StatusInternalServerError, fmt.Sprintf("Error uploading file: %s", err.Error()))
		return
	}

	// Set the permission for the new file
	permission := &drive.Permission{
		Type: "anyone",
		Role: "reader",
	}
	_, err = srv.Permissions.Create(res.Id, permission).Do()
	if err != nil {
		log.Fatalf("Unable to create permission: %v", err)
	}

	fmt.Printf("File '%s' uploaded successfully to parent folder '%s'.\n", res.Name, res.Id)

	ctx.JSON(200, gin.H{
		"message": "uploaded successfully",
		"res":     res,
	})
}
