package api

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

// gin-gonic middleware used to inject new postgres
// connection into request context
func PostgresSessionMiddleware() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        // create new persistence instance and connect to postgres
        db := NewPersistence(environConfig.Get("postgres_url"))
        conn, err := db.Connect()
        if err != nil {
            log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
        defer conn.Close()

        ctx.Set("persistence", db)
        ctx.Next()
    }
}