package mail_relay

import (
    "fmt"
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

// postgres middleware. note that the autoDefer variable
// is used to control if the database session should be
// closed on API return or not. Some
func PostgresSessionMiddleware(postgresUrl string) gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // create new persistence instance and connect to postgres
        db := NewPersistence(postgresUrl)
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