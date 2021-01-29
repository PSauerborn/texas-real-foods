package authenticator

import (
    "fmt"
    "context"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

var postgresURL string

type Authenticator struct{
    PostgresURL string
    Engine *gin.Engine
    ListenAddress string
    ListenPort int
}

func New(postgresUrl, listenAddress string, listenPort int) *Authenticator {
    // create new instance of authenticator
    router := gin.Default()
    router.Any("/authenticate", Authenticate)

    // set postgres url for global user
    postgresURL = postgresUrl
    return &Authenticator{
        PostgresURL: postgresUrl,
        Engine: router,
        ListenAddress: listenAddress,
        ListenPort: listenPort,
    }
}

func(auth *Authenticator) Run() {
    conn := fmt.Sprintf("%s:%d", auth.ListenAddress, auth.ListenPort)
    log.Info(fmt.Sprintf("starting new instance of authenticator at %s", conn))
    auth.Engine.Run(conn)
}

func Authenticate(ctx *gin.Context) {
    log.Info("received new authentication request")
    // retrieve API key from headers
    apiKey := ctx.Request.Header.Get("X-ApiKey")
    if len(apiKey) < 1 || apiKey == "undefined" {
        log.Warn("received request without API key")
        ctx.JSON(http.StatusUnauthorized,
            gin.H{"http_code": http.StatusUnauthorized, "message": "Unauthorized"})
        return
    }

    // create new persistence instance and connect to postgres
    db := NewPersistence(postgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    defer conn.Close(context.Background())

    // validate API key in postgres server
    valid, err := db.IsValidApiKey(apiKey)
    if err != nil {
        log.Error(fmt.Errorf("unable to validate API key: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }

    // return 200 is key is valid, else return 403
    if valid {
        log.Info("successfully validated request")
        ctx.JSON(http.StatusOK,
            gin.H{"http_code": http.StatusOK, "message": "Successfully authenticated"})
        return
    }
    log.Warn("received request with invalid API key")
    ctx.JSON(http.StatusForbidden,
        gin.H{"http_code": http.StatusForbidden, "message": "Forbidden"})
    return
}


