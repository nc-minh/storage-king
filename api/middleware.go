package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/nc-minh/storage-king/db/sqlc"
	"github.com/rs/zerolog/log"
)

func (server *Server) accessTokenMiddleware(ctx *gin.Context) {
	id := ctx.PostForm("id")
	email := ctx.PostForm("email")

	num, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		log.Err(err).Msg("unable to parse id")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.GetStorageParams{
		ID:    num,
		Email: email,
	}

	storage, err := server.store.GetStorage(ctx, arg)
	if err != nil {
		log.Err(err).Msg("unable to get storage")
		ctx.AbortWithStatusJSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	ctx.Set("access_token", storage.AccessToken)

	updatedAt := storage.UpdatedAt.Unix()
	accessTokenExpiresIn := storage.AccessTokenExpiresIn
	if !accessTokenExpiresIn.Valid {
		accessTokenExpiresIn.Int32 = 0
	}
	timeExpiresIn := updatedAt + int64(accessTokenExpiresIn.Int32)
	timeNow := time.Now().Unix()

	if timeExpiresIn < timeNow {
		log.Info().Msg("Access token expired, refreshing...")

		// Delete old client
		driveClient[id] = nil

		res, err := server.google.RefreshToken(storage.RefreshToken, server.config.ClientID, server.config.ClientSecret, storage.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, errorResponse(err))
			ctx.Abort()
			return
		}

		ctx.Set("access_token", res.AccessToken)
	}

	ctx.Next()
}

func (server *Server) authMiddleware(ctx *gin.Context) {
	cookie, err := ctx.Cookie("access_token")

	if err != nil {
		ctx.Redirect(http.StatusFound, "/authenticate")
	}

	ctx.Set("access_token", cookie)
	ctx.Next()
}
