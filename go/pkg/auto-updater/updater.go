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
func(updater *AutoUpdater) GetCurrentAssets() ([]connectors.BusinessInfo, error) {
    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
        return []connectors.BusinessInfo{}, err
    }
    defer conn.Close(context.Background())
    // get all assets from database and return
    return db.GetAllAssets()
}

// function used to process asset information
func(updater *AutoUpdater) ProcessAssets(assets []connectors.BusinessInfo) error {
    // establish new connection to postgres persistence
    db := NewPersistence(updater.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
        return err
    }
    defer conn.Close(context.Background())

    // iterate over assets and update
    for _, asset := range(assets) {
        if err := db.UpdateAsset(asset); err != nil {
            log.Error(fmt.Errorf("unable to update asset '%s': %+v", asset.BusinessName, err))
        }

        // generate new notification and send via engine
        notification := notifications.ChangeNotification{
            BusinessId: asset.BusinessId,
            BusinessName: asset.BusinessName,
            EventTimestamp: time.Now(),
            Notification: fmt.Sprintf("found new phone numbers %+v", asset.BusinessPhones),
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
    log.Info(fmt.Sprintf("starting new asset auto-updater with collection interval %d", updater.CollectionPeriodMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(updater.CollectionPeriodMinutes) * time.Second)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new collection job...")
                // retrieve current list of assets
                currentAssets, err := updater.GetCurrentAssets()
                if err != nil {
                    log.Error(fmt.Errorf("unable to retrieve existing assets: %+v", err))
                    continue
                }

                // retrieve updated asset list from connector
                assets, err := updater.DataConnector.CollectData(currentAssets)
                if err != nil {
                    log.Error(fmt.Errorf("unable to retrieve asset data: %+v", err))
                    continue
                }

                // process collected assets
                if len(assets) > 0 {
                    log.Info(fmt.Sprintf("successfully retrieved %d assets. processing...", len(assets)))
                    updater.ProcessAssets(assets)
                } else {
                    log.Info("no changes in assets detected. sleeping...")
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

