package connectors

import (
    "fmt"

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

func(connector *YelpAPIConnector) CollectData(businesses []connectors.BusinessMetadata) (
    []connectors.BusinessUpdate, error) {

    log.Info(fmt.Sprintf("updating data for %d assets", len(businesses)))
    updates := []connectors.BusinessUpdate{}

    for _, business := range(businesses) {
        // convert metadata field into struct. note that not all
        // business entries may have metadata fields required to
        // process YELP API requests
        meta, err := ParseYelpMetadata(business.Metadata)
        if err != nil {
            log.Warn(fmt.Sprintf("cannot process asset %s: invalid yelp metadata", business.BusinessId))
            continue
        }
        // collect new values for business and append to results
        updated, err := connector.UpdateBusiness(business, meta)
        if err != nil {
            log.Error(fmt.Errorf("unable to update asset: %+v", err))
            continue
        }
        updates = append(updates, updated)
    }
    return updates, nil
}

func(connector *YelpAPIConnector) UpdateBusiness(business connectors.BusinessMetadata,
    meta YelpMetadata) (connectors.BusinessUpdate, error) {
    // get business results form yelp API
    yelpResults, err := GetYelpBusinessInfo(meta.YelpBusinessId, connector.APIKey)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve asset data from yelp API: %+v", err))
        return connectors.BusinessUpdate{}, err
    }
    // generate new payload based on results from yelp API
    payload := connectors.BusinessData{
        WebsiteLive: true,
        BusinessPhones: []string{utils.CleanNumber(yelpResults.PhoneNumber)},
        Source: connector.Name(),
    }
    // generate new update and return
    update := connectors.BusinessUpdate{
        Meta: business,
        Data: payload,
    }
    return update, nil
}

func(connector *YelpAPIConnector) Name() string {
    return "yelp-api-connector"
}