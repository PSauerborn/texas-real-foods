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
    ErrAssetNotFound = errors.New("Cannot find specified asset")

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

    // add route to retrieve assets
    router.GET("/assets", PostgresSessionMiddleware(), getAssetsHandler)

    // add route to create new asset
    router.POST("/asset", PostgresSessionMiddleware(), addNewAssetHandler)

    // add routes to modify assets
    router.PATCH("/asset/:assetId", PostgresSessionMiddleware(), updateAssetHandler)
    router.DELETE("/asset/:assetId", PostgresSessionMiddleware(), deleteAssetHandler)
    return router
}

func getAssetsHandler(ctx *gin.Context) {
    log.Info("received request to retrieve assets")

    // retrieve postgres persistence from contex and
    db, _ := ctx.MustGet("persistence").(*Persistence)
    assets, err := db.GetAssets()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve assets: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "data": assets})
}

func addNewAssetHandler(ctx *gin.Context) {
    log.Info("received request create new asset")
    var request NewAssetRequest
    // parse request body and return 400 response if invalid body
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // retrieve postgres persistence from contex and add asset
    db, _ := ctx.MustGet("persistence").(*Persistence)
    if err := db.CreateAsset(request); err != nil {
        log.Error(fmt.Errorf("unable to generate new asset: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusCreated,
        gin.H{"http_code": http.StatusCreated, "message": "Successfully created asset"})
}

func updateAssetHandler(ctx *gin.Context) {
    log.Info("received request to update asset")
    var request NewAssetRequest
    // parse request body and return 400 response if invalid body
    if err := ctx.ShouldBind(&request); err != nil {
        log.Error(fmt.Errorf("unable to parse request body: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid request body"})
        return
    }

    // retrieve asset ID from parameters
    assetId, err := uuid.Parse(ctx.Param("assetId"))
    if err != nil {
        log.Error(fmt.Errorf("unable to parse parameter ID: %+v", err))
        ctx.JSON(http.StatusBadRequest,
            gin.H{"http_code": http.StatusBadRequest, "message": "Invalid asset ID"})
        return
    }

    // retrieve postgres persistence from contex
    db, _ := ctx.MustGet("persistence").(*Persistence)

    // check that asset exists in database
    _, err = db.GetAssetById(assetId)
    if err != nil {
        switch err {
        case ErrAssetNotFound:
            ctx.JSON(http.StatusNotFound,
                gin.H{"http_code": http.StatusNotFound, "message": "Invalid asset ID"})
            return
        default:
            ctx.JSON(http.StatusInternalServerError,
                gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
            return
        }
    }

    // update asset URI in database
    if err := db.UpdateAssetURI(request.BusinessURI, assetId); err != nil {
        log.Error(fmt.Errorf("unable to update asset asset: %+v", err))
        ctx.JSON(http.StatusInternalServerError,
            gin.H{"http_code": http.StatusInternalServerError, "message": "Internal server error"})
        return
    }
    ctx.JSON(http.StatusOK,
        gin.H{"http_code": http.StatusOK, "message": "Successfully updated asset"})
}

func deleteAssetHandler(ctx *gin.Context) {
    log.Info("received request to delete asset")
    ctx.JSON(http.StatusNotImplemented,
        gin.H{"http_code": http.StatusNotImplemented, "message": "Feature not yet available"})
}