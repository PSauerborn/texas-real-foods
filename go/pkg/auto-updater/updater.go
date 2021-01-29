package auto_updater

import (
    "fmt"
    "time"
    "sync"
    "context"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/notifications"
)

var (

)

// function used to create new auto updater
func New(connector connectors.AutoUpdateDataConnector, collectionPeriod int,
    notificationEngine notifications.NotificationEngine,
    postgresUrl string) *AutoUpdater {
    return &AutoUpdater{
        PostgresURL: postgresUrl,
        DataConnector: connector,
        CollectionPeriodMinutes: collectionPeriod,
        NotificationEngine: notificationEngine,
    }
}

type AutoUpdater struct{
    PostgresURL             string
    CollectionPeriodMinutes int
    DataConnector           connectors.AutoUpdateDataConnector
    NotificationEngine      notifications.NotificationEngine
}

// function used to process asset information
func(updater *AutoUpdater) GetCurrentBusinesses() ([]connectors.BusinessMetadata, error) {
    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
        return []connectors.BusinessMetadata{}, err
    }
    defer conn.Close(context.Background())
    // get all assets from database and return
    return db.GetAllMetadata()
}

// function used to process asset information
func(updater *AutoUpdater) ProcessBusinessUpdates(updates []connectors.BusinessUpdate) error {
    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
        return err
    }
    defer conn.Close(context.Background())

    // iterate over assets and update
    for _, update := range(updates) {
        if err := db.UpdateBusinessData(update); err != nil {
            log.Error(fmt.Errorf("unable to update business '%s': %+v", update.Meta.BusinessName, err))
        }

        // generate new notification and send via engine
        notification := notifications.ChangeNotification{
            BusinessId: update.Meta.BusinessId,
            BusinessName: update.Meta.BusinessName,
            EventTimestamp: time.Now(),
            Notification: fmt.Sprintf("found new phone numbers %+v", update.Data.BusinessPhones),
        }
        // send update notification and display errors
        err := updater.NotificationEngine.SendNotification(notification)
        if err != nil {
            log.Warn(fmt.Errorf("unable to send update notification: %+v", err))
        }
    }
    return nil
}

// function used to start updater. the auto updater
func(updater *AutoUpdater) Run() {
    log.Info(fmt.Sprintf("starting new business auto-updater with collection interval %d", updater.CollectionPeriodMinutes))
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
                // retrieve current list of businesses
                currentBusinesses, err := updater.GetCurrentBusinesses()
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

