package mail_relay

import (
    "fmt"
    "strings"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    apis "texas_real_foods/pkg/utils/api_accessors"
)

var (
    // define global
    mailChimpConfig MailChimpConfig
    // define global configuration for Utils API
    utilsAPIConfig utils.APIDependencyConfig

    persistence *Persistence
)

type MailChimpConfig struct {
    APIUrl string
    APIKey string
    ListID string
}


// function used to generate new mail
func NewMailRelay(cfg MailChimpConfig, utilsConfig utils.APIDependencyConfig,
    db *Persistence) *gin.Engine {
    // set MailChimp and Utils API config settings globally
    utilsAPIConfig = utilsConfig
    mailChimpConfig = cfg
    persistence = db

    router := gin.Default()
    // define health check route
    router.GET("/relay/health", healthCheckHandler)
    router.GET("/relay/history", relayHistoryHandler)

    // define route to relay mail service
    router.POST("/relay/mail-chimp", mailRelayHandler)
    return router
}

// function used to serve health check response and route
func healthCheckHandler(ctx *gin.Context) {
    log.Info("received request for health check route")
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "message": "Service running"})
}

// function used to relay mail service to mail chimp server
func mailRelayHandler(ctx *gin.Context) {
    log.Info("received new mail chimp relay request")
    var request MailRelayRequest
    // parse request body and validate
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // insert new mail entry into database
    entryId, err := persistence.InsertMailEntry(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to insert new mail entry into database: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    // set entry ID in request and send down channel to worker
    request.EntryId = entryId

    // get data for zip code (economic region) from API
    zipData, err := GetZipCodeData(request.ZipCode)
    if err != nil {
        persistence.UpdateMailEntry(entryId, "failed", false)
        switch err {
        case apis.ErrInvalidZipCode:
            ctx.JSON(http.StatusBadRequest, gin.H{"http_code": http.StatusBadRequest,
                "message": "Invalid zip code"})
        default:
            log.Error(fmt.Errorf("unable to fetch zip code data: %+v", err))
            ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
                "message": "Internal server error"})
        }
        return
    }

    async := ctx.DefaultQuery("async", "false")
    switch strings.ToLower(async) {
    case "true", "t":
        log.Info("executing new mail chimp request with async handler")
        go TriggerMailChimpAsync(request, zipData)
        ctx.JSON(http.StatusAccepted,
            gin.H{"http_code": http.StatusAccepted,
                  "message": "Successfully started mail relay job", "job_id": entryId})
    default:
        log.Info("executing new mail chimp request with sync handler")
        if err := TriggerMailChimp(request, zipData); err != nil {
            log.Error(fmt.Errorf("unable to execute mail chimp job: %+v", err))
            // insert failed message into event log
            persistence.UpdateMailEntry(entryId, "failed", false)
            ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
                "message": "Internal server error"})
        } else {
            // insert failed message into event log
            persistence.UpdateMailEntry(entryId, "completed", true)
            ctx.JSON(http.StatusOK, gin.H{"http_code": http.StatusOK,
                "message": "Successfully execute mail chimp job"})
        }
    }
}

// function used to relay mail service to mail chimp server
func relayHistoryHandler(ctx *gin.Context) {
    log.Info("received request to retreive history")

    data, err := persistence.GetMailEntries()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve mail entries from database: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": data})
}