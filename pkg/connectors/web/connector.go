package connectors

import (
    "fmt"
    "reflect"
    "errors"

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


func NewWebConnector() *WebConnector {
    return &WebConnector{}
}

type WebConnector struct{}

// function used to scrape sites for updated asset data
func(connector *WebConnector) Name() string {
    return "web-scraper"
}

// function used to collect data using webscraper
func(connector *WebConnector) CollectData(assets []connectors.BusinessInfo) ([]connectors.BusinessInfo, error) {
    log.Info(fmt.Sprintf("collecting data for %d assets using web connector", len(assets)))

    // generate new webscraper and save globally
    scraper = colly.NewCollector()
    scraper.AllowURLRevisit = false

    updatedAssets := []connectors.BusinessInfo{}
    // iterate over assets and scrape data
    for _, asset := range(assets) {
        log.Debug(fmt.Sprintf("scraping data for asset %+v", asset))
        // scrape site for updated asset information
        updated, err := connector.ScrapeSiteData(asset)
        if err != nil {
            log.Error(fmt.Sprintf("unable to scrape data for asset %+v: %+v", asset, err))
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

// function used to scrape sites for updated asset data
func(connector *WebConnector) ScrapeSiteData(asset connectors.BusinessInfo) (connectors.BusinessInfo, error) {

    var scrapeError error
    // add callbacks to scraper and start
    scraper.OnRequest(func(r *colly.Request) {
        log.Debug(fmt.Sprintf("making request to site %s", r.URL))
    })
    scraper.OnResponse(func(r *colly.Response) {
        log.Debug(fmt.Sprintf("connected to page with code %d", r.StatusCode))
        if r.StatusCode == 200 {
            // parse site and extract data
            asset, scrapeError = connector.ParseSiteData(asset, r.Body)
            if scrapeError != nil {
                log.Error(fmt.Errorf("unable to scrape site data: %+v", scrapeError))
                return
            }
        } else {
            // update asset to indicate that website is no longer active
            asset.WebsiteLive = false
        }
    })
    scraper.OnError(func(r *colly.Response, err error) {
        log.Error(fmt.Errorf("unable to scrape web data: %+v", err))
        scrapeError = err
    })

    // scrape website for data
    scraper.Visit(asset.BusinessURI)

    // handle any error raised during scraping of website
    if scrapeError != nil {
        switch scrapeError {
        default:
            return asset, scrapeError
        }
    }
    return asset, nil
}

func(connector *WebConnector) ParseSiteData(asset connectors.BusinessInfo,
    data []byte) (connectors.BusinessInfo, error) {

    log.Info(fmt.Sprintf("received and parsing %d bytes of data", len(data)))
    // parse site data for phone numbers by using regex expressions
    phones := utils.GetPhoneNumbersByRegex(string(data))
    asset.BusinessPhones = phones
    return asset, nil
}