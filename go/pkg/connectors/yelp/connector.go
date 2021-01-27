package connectors

import (
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/connectors"
	"texas_real_foods/pkg/utils"
)

type YelpAPIConnector struct{
	BaseAPIUrl string
	APIKey     string
}

func NewYelpAPIConnector(baseUrl, apiKey string) *YelpAPIConnector {
	return &YelpAPIConnector{
		BaseAPIUrl: baseUrl,
		APIKey: apiKey,
	}
}

func(connector *YelpAPIConnector) CollectData(assets []connectors.BusinessInfo) ([]connectors.BusinessInfo, error) {
	log.Info(fmt.Sprintf("updating data for %d assets", len(assets)))
	updatedAssets := []connectors.BusinessInfo{}

	for _, asset := range(assets) {

		// convert metadata field
		meta, err := ParseYelpMetadata(asset.Metadata)
		if err != nil {
			log.Warn(fmt.Sprintf("cannot process asset %s: invalid yelp metadata", asset.BusinessId))
			continue
		}

		updated, err := connector.UpdateAsset(asset, meta)
		if err != nil {
			log.Error(fmt.Errorf("unable to update asset: %+v", err))
			continue
		}

		// add updated asset to list of assets if values differ
        if !(reflect.DeepEqual(updated, asset)) {
            log.Debug("asset(s) differ... adding to list of updated")
            updatedAssets = append(updatedAssets, updated)
        }
	}
	return updatedAssets, nil
}

func(connector *YelpAPIConnector) UpdateAsset(asset connectors.BusinessInfo,
	meta YelpMetadata) (connectors.BusinessInfo, error) {
	// get business results form yelp API
	yelpResults, err := GetYelpBusinessInfo(meta.YelpBusinessId, connector.APIKey)
	if err != nil {
		log.Error(fmt.Errorf("unable to retrieve asset data from yelp API: %+v", err))
		return asset, err
	}
	// add updated phone number to asset
	asset.BusinessPhones = []string{
		utils.CleanNumber(yelpResults.PhoneNumber),
	}
	return asset, nil
}

func(connector *YelpAPIConnector) Name() string {
	return "yelp-api-connector"
}