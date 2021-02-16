package connectors

import (
    "fmt"
    "errors"
    "net/url"

    "github.com/gocolly/colly/v2"
    "github.com/PSauerborn/hermes/pkg/client"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
)

var (
    // collector used to scrape sites for data
    scraper *colly.Collector

    // define custom errors
    ErrInvalidURI = errors.New("Invalid URI")
)

// define function used to generate new web connector. note
// that each instance of the WebConnector is created with a
// Phone Validation API host, which is used to validate phone
// numbers scraped from a site(s)
func NewWebConnector(phoneApiHost string) *WebConnector {
    return &WebConnector{phoneApiHost}
}

// struct used to store
type WebConnector struct{
    PhoneValidationAPIHost string
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

    // generate new webscraper and save globally
    scraper = colly.NewCollector()
    scraper.AllowURLRevisit = false
    scraper.MaxDepth = 1
    scraper.IgnoreRobotsTxt = false
    scraper.Async = false

    updates := []connectors.BusinessUpdate{}
    // iterate over businesses and scrape data
    for _, business := range(businesses) {
        log.Debug(fmt.Sprintf("scraping data for business %+v", business))

        // parse url and extract host
        host, err := url.Parse(business.BusinessURI)
        if err != nil {
            log.Error(fmt.Errorf("unable to parse business URI: %+v", err))
            continue
        }
        // configure scraper to only have access to host domain
        scraper.AllowedDomains = []string{host.Host}

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

    var (scrapeError error; data connectors.BusinessData)
    // add callbacks to scraper and start
    scraper.OnRequest(func(r *colly.Request) {
        log.Debug(fmt.Sprintf("making request to site %s", r.URL))
    })

    // define response handler
    scraper.OnResponse(func(r *colly.Response) {
        log.Debug(fmt.Sprintf("connected to page with code %d", r.StatusCode))
        if r.StatusCode == 200 {
            // parse site and extract data
            data, scrapeError = connector.ParseSiteData(business, r.Body)
            if scrapeError != nil {
                log.Error(fmt.Errorf("unable to scrape site data: %+v", scrapeError))
                return
            }
        } else {
            // update asset to indicate that website is no longer active
            data = connectors.BusinessData{
                WebsiteLive: false,
                Source: connector.Name(),
                BusinessOpen: false,
            }
        }
    })
    // define error handler
    scraper.OnError(func(r *colly.Response, err error) {
        log.Error(fmt.Errorf("unable to scrape web data: %+v", err))
        scrapeError = err
    })

    // scrape website for data
    scraper.Visit(business.BusinessURI)
    scraper.Wait()

    // handle any error raised during scraping of website
    if scrapeError != nil {
        switch scrapeError {
        default:
            return connectors.BusinessUpdate{}, scrapeError
        }
    }
    // generate new business update and return
    update := connectors.BusinessUpdate{
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
    results, err := utils.ValidatePhoneNumbers(connector.PhoneValidationAPIHost, phones)
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