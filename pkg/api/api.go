package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)


func New() *gin.Engine {
	router := gin.Default()

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
	ctx.JSON(http.StatusNotImplemented,
		gin.H{"http_code": http.StatusNotImplemented, "message": "Feature not yet available"})
}

func deleteAssetHandler(ctx *gin.Context) {
	log.Info("received request to delete asset")
	ctx.JSON(http.StatusNotImplemented,
		gin.H{"http_code": http.StatusNotImplemented, "message": "Feature not yet available"})
}