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
    router.GET("/texas-real-foods/businesses", PostgresSessionMiddleware(),
        getBusinessesHandler)
    router.GET("/texas-real-foods/notifications", PostgresSessionMiddleware(),
        getNotificationsHandler)
    // add routes to retrieve static and timeseries data
    router.GET("/texas-real-foods/data/static/:businessId", PostgresSessionMiddleware(),
        getStaticDataHandler)
    router.GET("/texas-real-foods/data/timeseries/:businessId/:start/:end",
        PostgresSessionMiddleware(), getTimeSeriesHandler)

    // add route to create new business
    router.POST("/texas-real-foods/business", PostgresSessionMiddleware(), addNewBusinessHandler)

    // add routes to modify businesses
    router.PATCH("/texas-real-foods/business/info/:businessId", PostgresSessionMiddleware(),
        updateBusinessHandler)
    router.PATCH("/texas-real-foods/business/meta/:businessId", PostgresSessionMiddleware(),
        updateBusinessMetaHandler)

    router.DELETE("/texas-real-foods/business/:businessId", PostgresSessionMiddleware(),
        deleteBusinessHandler)
    return router
}

// API handler used to retrieve existing businesses
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

// API handler used to retrieve notifications
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

// API handler used to add new business
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

// API handler used to update business
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

// API handler used to update business metadata. metadata are
// updated via JSON patch operations performed in sequence on
// instance
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

// API handler used to delete a business from the database
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

// PI handler to retrieve static data from database
func getStaticDataHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request to retrieve static data for business %s", ctx.Param("businessId")))
    // retrieve business ID from parameters and convert to uuid
    businessId, err := uuid.Parse(ctx.Param("businessId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid business ID"})
        return
    }

    // retrieve persistence from context and check if business exists
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

    // get statis business data from database and return
    data, err := db.GetStaticBusinessData(businessId)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve business data: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": data})
}

// API handler used to retrieve timeseries data from database for
// a given business with business ID
func getTimeSeriesHandler(ctx *gin.Context) {
    log.Info(fmt.Sprintf("received request to retrieve timeseries data for business %s", ctx.Param("businessId")))
    // retrieve business ID from parameters and convert to uuid
    businessId, err := uuid.Parse(ctx.Param("businessId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid business ID"})
        return
    }

    timeRange, err := ParseTimeRange(ctx.Param("start"), ctx.Param("end"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse timestamps: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid time range"})
        return
    }

    // retrieve persistence from context and check if business exists
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

    // get statis business data from database and return
    data, err := db.GetTimeSeriesData(businessId, timeRange.Start, timeRange.End)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve business data: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": GroupTimeseriesDataBySource(data)})
}