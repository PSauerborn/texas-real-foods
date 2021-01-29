package connectors

import (
    "fmt"
    "errors"
    "net/url"

    "github.com/gocolly/colly/v2"
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


func NewWebConnector(phoneApiHost string) *WebConnector {
    return &WebConnector{phoneApiHost}
}

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

    // generate new webscraper and save globally
    scraper = colly.NewCollector()
    scraper.AllowURLRevisit = false
    scraper.MaxDepth = 1
    scraper.IgnoreRobotsTxt = false
    scraper.Async = false

    updates := []connectors.BusinessUpdate{}
    // iterate over assets and scrape data
    for _, business := range(businesses) {
        log.Debug(fmt.Sprintf("scraping data for asset %+v", business))

        // parse url and extract host
        host, err := url.Parse(business.BusinessURI)
        if err != nil {
            log.Error(fmt.Errorf("unable to parse business URI: %+v", err))
            continue
        }
        // configure scraper to only have access to host domain
        scraper.AllowedDomains = []string{host.Host}

        // scrape site for updated asset information
        update, err := connector.ScrapeSiteData(business)
        if err != nil {
            log.Error(fmt.Sprintf("unable to scrape data for business %+v: %+v", business, err))
            continue
        }
        updates = append(updates, update)
    }
    return updates, nil
}

// function used to scrape sites for updated asset data
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
    }
    return businessData, nil
}