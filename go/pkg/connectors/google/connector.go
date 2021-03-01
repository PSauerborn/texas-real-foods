package connectors

import (
    "fmt"

    "github.com/PSauerborn/hermes/pkg/client"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
)

type GoogleAPIConnector struct{
    BaseAPIUrl string
    APIKey     string
}

func NewGoogleAPIConnector(baseUrl, apiKey string) *GoogleAPIConnector {
    return &GoogleAPIConnector{
        BaseAPIUrl: baseUrl,
        APIKey: apiKey,
    }
}

func(connector *GoogleAPIConnector) CollectData(businesses []connectors.BusinessMetadata) (
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
        // process Google API requests
        meta, err := ParseGoogleMetadata(business.Metadata)
        if err != nil {
            log.Warn(fmt.Sprintf("cannot process business %s: invalid google metadata",
            business.BusinessId))
            continue
        }

        // increment hermes counter used to measure total number of sites scraped
        hermesClient.IncrementCounter("total_google_requests",
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

// function to collect business data from google place API
func(connector *GoogleAPIConnector) UpdateBusiness(business connectors.BusinessMetadata,
    meta GoogleMetadata) (connectors.BusinessUpdate, error) {
    log.Debug(fmt.Sprintf("collecting data from Google Place API for business %+v", business))
    // request data from google place API
    response, err := GetGoogleBusinessInfo(meta.GooglePlaceId, connector.APIKey)
    if err != nil {
        log.Error(fmt.Errorf("unable to collect data for business '%s': %+v", business.BusinessName, err))
        return connectors.BusinessUpdate{}, err
    }

    // generate new payload based on results from google API
    payload := connectors.BusinessData{
        WebsiteLive: true,
        BusinessOpen: response.BusinessStatus == "OPERATIONAL",
        BusinessPhones: []string{utils.CleanNumber(response.FormattedPhoneNumber)},
        Source: connector.Name(),
    }
    // generate new update and return
    update := connectors.BusinessUpdate{
        Meta: business,
        Data: payload,
    }
    return update, nil
}

func(connector *GoogleAPIConnector) Name() string {
    return "google-api-connector"
}
