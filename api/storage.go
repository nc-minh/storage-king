package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/rs/zerolog/log"
)

func (server *Server) CreateAuthURL(ctx *gin.Context) {
	authURL := server.google.CreateAuthURL()
	ctx.Redirect(302, authURL)
}

func (server *Server) CreateStorage(ctx *gin.Context) {
	code := ctx.Query("code")

	tok, err := server.google.CodeAuthentication(code)
	if err != nil {
		log.Err(err).Msg("error while authenticating code")
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	email, err := server.google.GetEmailFromAccessToken(tok.AccessToken)
	if err != nil {
		log.Err(err).Msg("error while getting email from access token")
		ctx.JSON(200, gin.H{
			"error": err.Error(),
		})
		return
	}

	arg := db.CreateStorageParams{
		Email:        email,
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
	}

	storage, err := server.store.CreateStorage(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, storage)
}

type getStorageRequest struct {
	ID    int64  `uri:"id" binding:"min=1"`
	Email string `uri:"id"`
}

func (server *Server) RefreshToken(ctx *gin.Context) {
	var req getStorageRequest

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetStorageParams{
		ID:    req.ID,
		Email: req.Email,
	}

	storage, err := server.store.GetStorage(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	server.google.RefreshToken(storage.RefreshToken, server.config.ClientID, server.config.ClientSecret)
}
