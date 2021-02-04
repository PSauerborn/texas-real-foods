package mail_relay

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

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

    db, _ := ctx.MustGet("persistence").(*Persistence)
    // insert new mail entry into database
    entryId, err := db.InsertMailEntry(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to insert new mail entry into database: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    // set entry ID in request and send down channel to worker
    request.EntryId = entryId
    eventChannel <- request

    ctx.JSON(http.StatusAccepted,
        gin.H{
            "http_code": http.StatusAccepted,
            "message": "Successfully started mail relay job",
            "job_id": entryId})
}

// function used to relay mail service to mail chimp server
func relayHistoryHandler(ctx *gin.Context) {
    log.Info("received request to retreive history")

    db, _ := ctx.MustGet("persistence").(*Persistence)
    data, err := db.GetMailEntries()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve mail entries from database: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }

    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": data})
}