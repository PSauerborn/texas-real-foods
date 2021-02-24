package notifications

import (
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)

// function to generate a new gin router with the
// relevant routes set
func NewNotificationService(postgresUrl string) *gin.Engine {
    router := gin.Default()
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowCredentials: true,
        AllowHeaders:     []string{"*"},
    }))

    // add postgres session middleware to API
    router.Use(PostgresSessionMiddleware(postgresUrl))

    router.GET("/notifications/all", getNotificationsHandler)
    router.GET("/notifications/unread", getUnreadNotificationsHandler)

    router.POST("/notifications/new", createNotificationHandler)

    router.PATCH("/notifications/update/:notificationId", updateNotificationHandler)
    return router
}

// API handler used to retrieve notification
func getNotificationsHandler(ctx *gin.Context) {
    log.Info("received request to retrieve notifications")
    // retrieve persistence layer from context and retrieve
    // all current notifications
    db, _ := ctx.MustGet("persistence").(*Persistence)
    notifications, err := db.GetNotifications()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve notifications: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"http_code": http.StatusOK,
        "count": len(notifications), "notifications": notifications})
}

// API handler used to retrieve unread notifications
func getUnreadNotificationsHandler(ctx *gin.Context) {
    log.Info("received request to retrieve unread notifications")
    // retrieve persistence layer from context and retrieve
    // all unread notifications
    db, _ := ctx.MustGet("persistence").(*Persistence)
    notifications, err := db.GetUnreadNotifications()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve notifications: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"http_code": http.StatusOK,
        "count": len(notifications), "notifications": notifications})
}

// API handler used to update notifications
func updateNotificationHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request to update notification %s", ctx.Param("notificationId")))
    // extract notification ID from path and parse to UUID
    notificationId, err := uuid.Parse(ctx.Param("notificationId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse notification ID: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{"http_code": http.StatusBadRequest,
            "message": "Invalid notification ID"})
        return
    }
    // check if notification exists or not
    db, _ := ctx.MustGet("persistence").(*Persistence)
    exists, err := db.NotificationExists(notificationId)
    if err != nil {
        log.Error(fmt.Errorf("unable to check existing notifications: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    } else if !exists {
        ctx.JSON(http.StatusNotFound, gin.H{"http_code": http.StatusNotFound,
            "message": "Invalid notification ID"})
        return
    }

    // update notification in database
    if err := db.UpdateNotification(notificationId); err != nil {
        log.Error(fmt.Errorf("unable to update notification: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"http_code": http.StatusOK,
        "message": "Successfully updated notification"})
}

// API handler used to create a new notification
func createNotificationHandler(ctx *gin.Context) {
    log.Info("received request to generate new notification")
    var request ChangeNotification
    // parse request body and validate
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest, gin.H{"http_code": http.StatusBadRequest,
            "message": "Invalid request body"})
        return
    }

    // extract persistence from context and create new notification
    db, _ := ctx.MustGet("persistence").(*Persistence)

    exists, err := db.NotificationHashExists(request.NotificationHash)
    if err != nil {
        log.Error(fmt.Errorf("unable to validate notification hash: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    } else if exists {
        log.Error(fmt.Errorf("unable to validate notification hash: %+v", err))
        ctx.JSON(http.StatusConflict, gin.H{"http_code": http.StatusConflict,
            "message": "Notification hash already exists"})
        return
    }

    if err := db.CreateNotification(request); err != nil {
        log.Error(fmt.Errorf("unable to create notification: %+v", err))
        ctx.JSON(http.StatusInternalServerError, gin.H{"http_code": http.StatusInternalServerError,
            "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusCreated, gin.H{"http_code": http.StatusCreated,
        "message": "Successfully created notification"})
}



