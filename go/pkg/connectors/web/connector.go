package connectors

import (
    "fmt"
    "errors"
    "net/http"
    "io/ioutil"

    "github.com/PSauerborn/hermes/pkg/client"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
    api "texas_real_foods/pkg/utils/api_accessors"
)

var (
    // define custom errors
    ErrInvalidURI = errors.New("Invalid URI")
)

// define function used to generate new web connector. note
// that each instance of the WebConnector is created with a
// Phone Validation API host, which is used to validate phone
// numbers scraped from a site(s)
func NewWebConnector(apiConfig utils.APIDependencyConfig) *WebConnector {
    return &WebConnector{apiConfig}
}

// struct used to store
type WebConnector struct{
    UtilsAPIConfig utils.APIDependencyConfig
}

// function used to scrape sites for updated asset
func(connector *WebConnector) Name() string {
    return "web-scraper"
}

// function used to collect data using webscraper
func(connector *WebConnector) CollectData(businesses []connectors.BusinessMetadata) (
    []connectors.BusinessUpdate, error) {
    log.Info(fmt.Sprintf("collecting data for %d businesses using web connector", len(businesses)))

    // create new instance of hermes client to update prometheus metrics
    hermesClient := hermes_client.New("texas-real-foods-hermes", 7789)
    labels := map[string]string{"source": connector.Name()}
    // increment gauge measuring running jobs and defer decrementing
    hermesClient.IncrementGauge("running_collection_jobs", labels)
    defer hermesClient.DecrementGauge("running_collection_jobs", labels)

    updates := []connectors.BusinessUpdate{}
    // iterate over businesses and scrape data
    for _, business := range(businesses) {
        log.Debug(fmt.Sprintf("scraping data for business %+v", business))
        // scrape site for updated business information
        update, err := connector.ScrapeSiteData(business)
        // increment hermes counter used to measure total number of sites scraped
        hermesClient.IncrementCounter("total_sites_scraped",
            map[string]string{"business_name": business.BusinessName})
        if err != nil {
            log.Error(fmt.Sprintf("unable to scrape data for business %+v: %+v", business, err))
            continue
        }
        updates = append(updates, update)
    }
    return updates, nil
}

// function used to scrape sites for updated business data
func(connector *WebConnector) ScrapeSiteData(business connectors.BusinessMetadata) (
    connectors.BusinessUpdate, error) {

    var (scrapeError error; data connectors.BusinessData; update connectors.BusinessUpdate)
    // add callbacks to scraper and start
    // generate new HTTP request with given settings
    req, err := http.NewRequest("GET", business.BusinessURI, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP Request: %+v", err))
        return update, err
    }

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return update, err
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case 200:
        // extract request body
        bytes, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Error(fmt.Errorf("unable to read response body: %+v", err))
            return update, err
        }
        // parse site and extract data
        data, scrapeError = connector.ParseSiteData(business, bytes)
        if scrapeError != nil {
            log.Error(fmt.Errorf("unable to scrape site data: %+v", scrapeError))
        }
    default:
        log.Error(fmt.Errorf("unable to scrape site data: received status code %d", resp.StatusCode))
        // update asset to indicate that website is no longer active
        data = connectors.BusinessData{
            WebsiteLive: false,
            Source: connector.Name(),
            BusinessOpen: false,
        }
    }
    // generate new business update and return
    update = connectors.BusinessUpdate{
        Meta: business,
        Data: data,
    }
    return update, nil
}

// function used to parse data downloaded from website
func(connector *WebConnector) ParseSiteData(business connectors.BusinessMetadata,
    data []byte) (connectors.BusinessData, error) {

    log.Info(fmt.Sprintf("received and parsing %d bytes of data", len(data)))
    // parse site data for phone numbers by using regex expressions
    phones := utils.GetPhoneNumbersByRegex(string(data))

    // create new accessor for utils API and validate phone numbers
    access := api.NewUtilsAPIAccessor(connector.UtilsAPIConfig.Host, "http",
        connector.UtilsAPIConfig.Port)
    results, err := access.ValidatePhoneNumbers(phones)
    if err != nil {
        log.Error(fmt.Errorf("unable to verify phone numbers with API: %+v", err))
        return connectors.BusinessData{}, err
    }
    log.Debug(fmt.Sprintf("Phone API returned response %+v", results))
    // assign valid phone numbers to asset
    businessData := connectors.BusinessData{
        WebsiteLive: true,
        BusinessPhones: phones,
        Source: connector.Name(),
        BusinessOpen: true,
    }
    return businessData, nil
}