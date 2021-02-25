package auto_updater

import (
    "fmt"
    "time"
    "sync"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
    api "texas_real_foods/pkg/utils/api_accessors"
)


// function used to create new auto updater. the auto-updater
// takes a variety of components to operate. particularly
// important is the instance of a AutoUpdateDataConnector
// interface implementation, which is used to collect data from
// a particular data source (yelp, google, website etc)
func New(connector connectors.AutoUpdateDataConnector, collectionPeriod int,
    postgresUrl string, apiConfig utils.APIDependencyConfig) *AutoUpdater {
    return &AutoUpdater{
        PostgresURL: postgresUrl,
        DataConnector: connector,
        CollectionPeriodMinutes: collectionPeriod,
        TRFApiConfig: apiConfig,
    }
}

// struct to store components for auto updater. note that each
// instance has a separate data connector and notification engine
type AutoUpdater struct{
    PostgresURL             string
    CollectionPeriodMinutes int
    TRFApiConfig            utils.APIDependencyConfig
    DataConnector           connectors.AutoUpdateDataConnector
    StreamedConnector       connectors.StreamedAutoUpdateDataConnector
}

// function used to retrieve business metadata for all stored
// businesses
func(updater *AutoUpdater) GetCurrentBusinesses(host string, port *int) ([]connectors.BusinessMetadata, error) {
    // establish new connection to postgres persistence
    access := api.NewTRFApiAccessor(host, "http", port)
    payload, err := access.GetBusinesses()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve businesses from API: %+v", err))
        return payload.Data, err
    }
    return payload.Data, nil
}

// function used to process business data updates
func(updater *AutoUpdater) ProcessBusinessUpdates(updates []connectors.BusinessUpdate) error {
    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
        return err
    }
    defer conn.Close()

    // iterate over businesses and update in database
    for _, update := range(updates) {
        if err := db.UpdateBusinessData(update); err != nil {
            log.Warn(fmt.Errorf("unable to update business '%s': %+v",
                update.Meta.BusinessName, err))
        }
    }
    return nil
}

// function used to start updater. the auto updater
func(updater *AutoUpdater) Run() {
    log.Info(fmt.Sprintf("starting new business auto-updater with collection interval %d",
    updater.CollectionPeriodMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(updater.CollectionPeriodMinutes) * time.Minute)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new collection job...")
                start := time.Now()
                // retrieve current list of businesses
                currentBusinesses, err := updater.GetCurrentBusinesses(updater.TRFApiConfig.Host,
                    updater.TRFApiConfig.Port)
                if err != nil {
                    log.Error(fmt.Errorf("unable to retrieve existing businesses: %+v", err))
                    continue
                }

                // retrieve updated asset list from connector
                updates, err := updater.DataConnector.CollectData(currentBusinesses)
                if err != nil {
                    log.Error(fmt.Errorf("unable to retrieve business data: %+v", err))
                    continue
                }

                // process collected business updates
                if len(updates) > 0 {
                    log.Info(fmt.Sprintf("successfully retrieved %d updates. processing...", len(updates)))
                    updater.ProcessBusinessUpdates(updates)
                } else {
                    log.Info("no changes in business data detected. sleeping...")
                }
                // log total time elapsed to process job
                elapsed := time.Now().Sub(start)
                log.Info(fmt.Sprintf("finished update job. took %fs to process", elapsed.Seconds()))
            case <- quitChan:
                // stop ticker and add to waitgroup
                ticker.Stop()
                wg.Done()
                return
            }
        }
    }()

    wg.Wait()
    log.Info("stopping auto-updater...")
}

