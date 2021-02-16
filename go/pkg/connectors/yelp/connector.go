package connectors

import (
    "fmt"

    "github.com/PSauerborn/hermes/pkg/client"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
)

// define struct for Yelp API connector
type YelpAPIConnector struct{
    // base API url to use for YELP Fusion API
    BaseAPIUrl string
    // API key to gain access to API
    APIKey     string
}

// function used to generate a new Yelp API connector. note that
// both the base URL and a valid API key must be provided to the
// constructor
func NewYelpAPIConnector(baseUrl, apiKey string) *YelpAPIConnector {
    return &YelpAPIConnector{
        BaseAPIUrl: baseUrl,
        APIKey: apiKey,
    }
}

// function used to collect data from YELP API. specific keys of
// interest are extracted from the API response and send to the
// updater to be stored in the postgres
func(connector *YelpAPIConnector) CollectData(businesses []connectors.BusinessMetadata) (
    []connectors.BusinessUpdate, error) {

    log.Info(fmt.Sprintf("updating data for %d businesses", len(businesses)))
    updates := []connectors.BusinessUpdate{}

    // create new instance of hermes client to update prometheus metrics
    hermesClient := hermes_client.New("texas-real-foods-hermes", 7789)
    labels := map[string]string{"source": connector.Name()}
    // increment gauge measuring running jobs and defer decrementing
    hermesClient.IncrementGauge("running_collection_jobs", labels)
    defer hermesClient.DecrementGauge("running_collection_jobs", labels)

    for _, business := range(businesses) {
        // convert metadata field into struct. note that not all
        // business entries may have metadata fields required to
        // process YELP API requests
        meta, err := ParseYelpMetadata(business.Metadata)
        if err != nil {
            log.Warn(fmt.Sprintf("cannot process business %s: invalid yelp metadata",
            business.BusinessId))
            continue
        }

        // increment hermes counter used to measure total number of sites scraped
        hermesClient.IncrementCounter("total_yelp_requests",
            map[string]string{"business_name": business.BusinessName})
        // collect new values for business and append to results
        updated, err := connector.UpdateBusiness(business, meta)
        if err != nil {
            log.Error(fmt.Errorf("unable to update business: %+v", err))
            continue
        }
        updates = append(updates, updated)
    }
    return updates, nil
}

// function used to call Yelp API to update business data
func(connector *YelpAPIConnector) UpdateBusiness(business connectors.BusinessMetadata,
    meta YelpMetadata) (connectors.BusinessUpdate, error) {
    // get business results form yelp API
    yelpResults, err := GetYelpBusinessInfo(meta.YelpBusinessId, connector.APIKey)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve business data from yelp API: %+v", err))
        return connectors.BusinessUpdate{}, err
    }
    // generate new payload based on results from yelp API
    payload := connectors.BusinessData{
        WebsiteLive: true,
        BusinessPhones: []string{utils.CleanNumber(yelpResults.PhoneNumber)},
        Source: connector.Name(),
        BusinessOpen: yelpResults.IsOpen,
    }
    // generate new update and return
    update := connectors.BusinessUpdate{
        Meta: business,
        Data: payload,
    }
    return update, nil
}

// function used to return source name from connector
func(connector *YelpAPIConnector) Name() string {
    return "yelp-api-connector"
}