package api

import (
    "fmt"
    "net/http"
    "errors"

    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/cors"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
)


var (
    // define custom errors
    ErrBusinessNotFound = errors.New("Cannot find specified business")

    // create map to house environment variables
    environConfig = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
        },
    )
)


func New() *gin.Engine {
    // create new gin router and add cors middleware
    router := gin.Default()
    router.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowCredentials: true,
    }))

    // add route to retrieve businesses
    router.GET("/texas-real-foods/businesses", PostgresSessionMiddleware(), getBusinessesHandler)
    router.GET("/texas-real-foods/notifications", PostgresSessionMiddleware(), getNotificationsHandler)

    // add route to create new business
    router.POST("/texas-real-foods/business", PostgresSessionMiddleware(), addNewBusinessHandler)

    // add routes to modify businesses
    router.PATCH("/texas-real-foods/business/info/:businessId", PostgresSessionMiddleware(), updateBusinessHandler)
    router.PATCH("/texas-real-foods/business/meta/:businessId", PostgresSessionMiddleware(), updateBusinessMetaHandler)

    router.DELETE("/texas-real-foods/business/:businessId", PostgresSessionMiddleware(), deleteBusinessHandler)
    return router
}

func getBusinessesHandler(ctx *gin.Context) {
    log.Info("received request to retrieve businesses")

    // retrieve postgres persistence from contex and
    db, _ := ctx.MustGet("persistence").(*Persistence)
    businesses, err := db.GetBusinesses()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve businesses: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    log.Info(fmt.Sprintf("retrieved %d entries from database", len(businesses)))
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": businesses})
}

func getNotificationsHandler(ctx *gin.Context) {
    log.Info("received request to retrieve notifications")

    // retrieve postgres persistence from contex and
    db, _ := ctx.MustGet("persistence").(*Persistence)
    notifications, err := db.GetNotifications()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve notifications: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    log.Info(fmt.Sprintf("retrieved %d entries from database", len(notifications)))
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": notifications})
}

func addNewBusinessHandler(ctx *gin.Context) {
    log.Info("received request create new business")
    var request NewBusinessRequest
    // parse request body and return 400 response if invalid body
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // retrieve postgres persistence from contex and add business
    db, _ := ctx.MustGet("persistence").(*Persistence)
    if err := db.CreateBusiness(request); err != nil {
        log.Error(fmt.Errorf("unable to generate new business: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusCreated,
        gin.H{"http_code": http.StatusCreated, "message": "Successfully created business"})
}

func updateBusinessHandler(ctx *gin.Context) {
    log.Info("received request to update business")
    var request BusinessUpdateRequest
    // parse request body and return 400 response if invalid body
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }
    // retrieve business ID from parameters
    businessId, err := uuid.Parse(ctx.Param("businessId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid business ID"})
        return
    }
    // retrieve postgres persistence from contex
    db, _ := ctx.MustGet("persistence").(*Persistence)
    _, err = db.GetBusinessById(businessId)
    if err != nil {
        switch err {
        case ErrBusinessNotFound:
            ctx.JSON(http.StatusNotFound,
                gin.H{"http_code": http.StatusNotFound, "message": "Invalid business ID"})
            return
        default:
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
    }
    // update business URI in database
    if err := db.UpdateBusinessURI(request.BusinessURI, businessId); err != nil {
        log.Error(fmt.Errorf("unable to update business: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "message": "Successfully updated business"})
}

func updateBusinessMetaHandler(ctx *gin.Context) {
    log.Info("received request to update business metadata")
    var request BusinessMetaPatchRequest
    // parse request body and return 400 response if invalid body
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // retrieve business ID from parameters
    businessId, err := uuid.Parse(ctx.Param("businessId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid business ID"})
        return
    }

    // retrieve postgres persistence from contex and retrieve business from database
    db, _ := ctx.MustGet("persistence").(*Persistence)
    business, err := db.GetBusinessById(businessId)
    if err != nil {
        switch err {
        case ErrBusinessNotFound:
            ctx.JSON(http.StatusNotFound,
                gin.H{"http_code": http.StatusNotFound, "message": "Invalid business ID"})
            return
        default:
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
    }

    // apply JSON patch to metadata object
    modified, err := PatchBusinessMeta(business, request.Operation)
    if err != nil {
        log.Error(fmt.Errorf("unable to apply JSON patch: %+v", err))
        switch err {
        case ErrInvalidPatch:
            ctx.JSON(http.StatusBadRequest,
                gin.H{"http_code": http.StatusBadRequest, "message": "Invalid JSON patch operation"})
            return
        case ErrInvalidBusinessMeta:
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
    }

    // update patched metadata in database
    if err := db.UpdateBusinessMetadata(modified, businessId); err != nil {
        log.Error(fmt.Errorf("unable to update business: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "message": "Successfully updated business"})
}

func deleteBusinessHandler(ctx *gin.Context) {
    log.Info("received request to delete business")
    // retrieve business ID from parameters
    businessId, err := uuid.Parse(ctx.Param("businessId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid business ID"})
        return
    }
    // retrieve postgres persistence from contex
    db, _ := ctx.MustGet("persistence").(*Persistence)
    _, err = db.GetBusinessById(businessId)
    if err != nil {
        switch err {
        case ErrBusinessNotFound:
            ctx.JSON(http.StatusNotFound,
                gin.H{"http_code": http.StatusNotFound, "message": "Invalid business ID"})
            return
        default:
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
    }

    if err := db.DeleteBusiness(businessId); err != nil {
        log.Error(fmt.Errorf("unable to delete business: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "message": "Successfully deleted business"})
}