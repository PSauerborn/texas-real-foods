package auto_updater

import (
    "fmt"
    "time"
    "sync"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
)

// function used to create new streaming auto updater. the auto-updater
// takes a variety of components to operate. particularly
// important is the instance of a AutoUpdateDataConnector
// interface implementation, which is used to collect data from
// a particular data source (yelp, google, website etc)
func NewStreamedAutoUpdater(connector connectors.StreamedAutoUpdateDataConnector,
    collectionPeriod int, postgresUrl string, apiConfig utils.APIDependencyConfig) *AutoUpdater {
    return &AutoUpdater{
        PostgresURL: postgresUrl,
        StreamedConnector: connector,
        CollectionPeriodMinutes: collectionPeriod,
        TRFApiConfig: apiConfig,
    }
}

// function used to process a single update for a given business
func(updater *AutoUpdater) ProcessSingleBusinessUpdate(db *Persistence,
    update connectors.BusinessUpdate) error {
    // iterate over businesses and update in database
    if err := db.UpdateBusinessData(update); err != nil {
        log.Warn(fmt.Errorf("unable to update business '%s': %+v",
            update.Meta.BusinessName, err))
    }
    return nil
}

// function used to start updater. the auto updater
func(updater *AutoUpdater) RunWithStreaming() {
    log.Info(fmt.Sprintf("starting new business auto-updater with collection interval %d",
        updater.CollectionPeriodMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(updater.CollectionPeriodMinutes) * time.Minute)
    quitChan := make(chan bool)

    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
        panic(err)
    }
    defer conn.Close()

    // generate new event queue to process business update
    updates := make(chan connectors.BusinessUpdate)
    go func() {
        for update := range(updates) {
            if err := updater.ProcessSingleBusinessUpdate(db, update); err != nil {
                log.Error(fmt.Errorf("unable to updated business %s: %+v",
                    update.Meta.BusinessName, err))
            }
        }
    }()

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

                log.Debug(fmt.Sprintf("scraping business data for %d businesses...", len(currentBusinesses)))
                // retrieve updated asset list from connector
                if err := updater.StreamedConnector.StreamData(updates, currentBusinesses); err != nil {
                    log.Error(fmt.Errorf("unable to retrieve business data: %+v", err))
                    continue
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
