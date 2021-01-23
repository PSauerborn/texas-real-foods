package api

import (
	"fmt"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	postgresURL = "postgres://postgres:postgres-dev@192.168.99.100:5432"
)

func PostgresSessionMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// create new persistence instance and connect to postgres
		db := NewPersistence(postgresURL)
		conn, err := db.Connect()
		if err != nil {
			log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
			ctx.JSON(http.StatusInternalServerError,
				gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
			return
		}
		defer conn.Close(context.Background())

		ctx.Set("persistence", db)
		ctx.Next()
	}
}