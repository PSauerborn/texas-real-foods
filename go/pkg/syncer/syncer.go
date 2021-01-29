package syncer

import (
    "fmt"
    "time"
    "sync"
    "errors"
    "context"
    "reflect"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/notifications"
    "texas_real_foods/pkg/connectors"
)

var (
    ErrDataSyncError = errors.New("Error syncing data tables")
)

type Syncer struct{
    PostgresURL   string
    Notifications notifications.NotificationEngine
    CollectionPeriodMinutes int
}

func NewSyncer(postgresUrl string, collectionPeriodMinutes int,
    notifier notifications.NotificationEngine) *Syncer {
    return &Syncer{
        PostgresURL: postgresUrl,
        Notifications: notifier,
        CollectionPeriodMinutes: collectionPeriodMinutes,
    }
}

func(syncer *Syncer) SyncData() error {
    // establish new connection to postgres persistence
    db := NewPersistence(syncer.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect to postgres: %+v", err))
        return err
    }
    defer conn.Close(context.Background())
    // get all businesses from database and return
    businesses, err := db.GetAllMetadata()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve business metadata: %+v", err))
        return err
    }

    // iterate over retrieved businesses and compare data values
    for _, business := range(businesses) {
        log.Debug(fmt.Sprintf("processing changes for business %+v", business))
        // retrieve all data values/source data for a given business
        data, err := db.GetDataByBusinessId(business.BusinessId)
        if err != nil {
            log.Error(fmt.Errorf("unable to retrieve data for business %s: %+v",
            business.BusinessId, err))
            continue
        }

        log.Debug(fmt.Sprintf("comparing %d entries for %s", len(data), business.BusinessName))
        if syncer.DataEntriesDiffer(data) {
            log.Info(fmt.Sprintf("found data differences in updates. sending notification..."))

            // convert entries into JSON string and hash. notifications are all
            // stored against a hash of the values that generated the notifications
            // to ensure that notifications from a given set of data values are
            // only generated once
            mapValues := map[string][]connectors.BusinessUpdate{ "entries": data }
            hashed, err := HashMap(mapValues)
            if err != nil {
                log.Error(fmt.Errorf("unable to generate hash from business update(s): %+v", err))
                continue
            }

            log.Debug(fmt.Sprintf("successfully hashed business update(s) into '%s'", hashed))

            // construt message for notification
            sources := []string{}
            for _, entry := range(data) {
                sources = append(sources, entry.Data.Source)
            }
            message := fmt.Sprintf("found differences in data sources %+v for business %s",
            sources, business.BusinessName)

            // send notification
            if err := syncer.SendNotification(message, hashed, business); err != nil {
                log.Error(fmt.Errorf("unable to send change notification: %+v", err))
                continue
            }
        } else {
            log.Info(fmt.Sprintf("data entries for '%s' in sync", business.BusinessName))
        }
    }
    return nil
}

// function to check if data entries differ
func(syncer *Syncer) DataEntriesDiffer(entries []connectors.BusinessUpdate) bool {

    mappedValues := map[string]ReducedBusinessData{}
    for _, entry := range(entries) {
        // get source of data and add to map
        source := entry.Data.Source
        mappedValues[source] = ReducedBusinessData{
            BusinessPhones: entry.Data.BusinessPhones,
            WebsiteLive: entry.Data.WebsiteLive,
        }
    }

    // compare values by source
    for source, data  := range(mappedValues) {
        for subSource, subData  := range(mappedValues) {
            if source == subSource {
                continue
            }
            if !reflect.DeepEqual(data, subData) {
                return true
            }
        }
    }
    return false
}

func(syncer *Syncer) SendNotification(message, hashed string,
    business connectors.BusinessMetadata) error {

    log.Debug(fmt.Sprintf("sending new message '%s'", message))
    payload := notifications.ChangeNotification{
        BusinessId: business.BusinessId,
        BusinessName: business.BusinessName,
        EventTimestamp: time.Now(),
        JSONHash: hashed,
        Notification: message,
    }
    return syncer.Notifications.SendNotification(payload)
}

func(syncer *Syncer) Run() {
    log.Info(fmt.Sprintf("starting new business syncer with collection interval %d",
    syncer.CollectionPeriodMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(syncer.CollectionPeriodMinutes) * time.Minute)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new sync job...")
                if err := syncer.SyncData(); err != nil {
                    log.Error(fmt.Errorf("unable to sync data: %+v", err))
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
