package timeseries_analyser

import (
    "fmt"
    "time"
    "sync"

    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/notifications"
    "texas_real_foods/pkg/utils"
    api "texas_real_foods/pkg/utils/api_accessors"
    "texas_real_foods/pkg/connectors"
)

type TimeseriesAnalyser struct {
    TRFAPIConfig  utils.APIDependencyConfig
    NotifyAPIConfig  utils.APIDependencyConfig
    AnalysisIntervalMinutes int
}

func NewAnalyser(apiConfig utils.APIDependencyConfig,
    notifyApiConfig utils.APIDependencyConfig,
    interval int) *TimeseriesAnalyser {
    return &TimeseriesAnalyser{
        TRFAPIConfig: apiConfig,
        NotifyAPIConfig: notifyApiConfig,
        AnalysisIntervalMinutes: interval,
    }
}

// function used to retrieve business metadata for all stored
// businesses
func(analyser *TimeseriesAnalyser) GetCurrentBusinesses(config utils.APIDependencyConfig) (
    []connectors.BusinessMetadata, error) {
    // establish new connection to postgres persistence
    access := api.NewTRFApiAccessorFromConfig(config)
    payload, err := access.GetBusinesses()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve businesses from API: %+v", err))
        return payload.Data, err
    }
    return payload.Data, nil
}

// function used to retrieve business metadata for all stored
// businesses
func(analyser *TimeseriesAnalyser) GetTimeseriesData(config utils.APIDependencyConfig,
    businessId uuid.UUID, start, end time.Time) (map[string][]api.TimeseriesDataEntry, error) {
    // establish new connection to postgres persistence
    access := api.NewTRFApiAccessorFromConfig(config)
    payload, err := access.GetTimeseriesDataCounted(businessId, 5)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve businesses from API: %+v", err))
        return payload.Data, err
    }
    return payload.Data, nil
}

// function used to analyse business data with a given
// business ID, start and end timestamps
func(analyser *TimeseriesAnalyser) AnalyseBusinessData(business connectors.BusinessMetadata,
    start, end time.Time) error {

    // get timeseries data from texas real foods API
    data, err := analyser.GetTimeseriesData(analyser.TRFAPIConfig,
        business.BusinessId, start, end)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve timeseries data for business %s: %+v",
            business.BusinessId, err))
        return err
    }

    notify := []notifications.ChangeNotification{}
    // iterate over source data and compare values
    for source, values := range(data) {
        log.Debug(fmt.Sprintf("analysing source %s with values %+v", source, values))
        for i, entry := range(values) {
            if i == 0 {
                continue
            }
            // compare timeseries entry to previous entry
            changed, fields := timeSeriesEntriesDiffer(entry, values[i - 1])
            if changed {
                log.Info(fmt.Sprintf("found differences in timeseries entries for %+v:%s",
                    business.BusinessName, source))
                // generate notification message and hash
                notificationString := fmt.Sprintf("Found change in timeseries business data for %s in source %s",
                    business.BusinessName, source)
                notificationString = fmt.Sprintf("%s: the following fields have changed %+v",
                    notificationString, fields)

                // generate new notification
                notification := notifications.ChangeNotification{
                    BusinessId: business.BusinessId,
                    BusinessName: business.BusinessName,
                    EventTimestamp: time.Now(),
                    Notification: notificationString,
                    NotificationHash: generateNotificationHash(business.BusinessId,
                        source, entry.EventTimestamp),
                    Metadata: map[string]interface{}{
                        "source": source,
                    },
                }
                notify = append(notify, notification)
                break
            }
        }
    }

    accessor := api.NewNotificationsApiAccessorFromConfig(analyser.NotifyAPIConfig)
    for _, msg := range(notify) {
        // send notification for business change
        if _, err := accessor.CreateNotification(msg); err != nil {
            log.Error(fmt.Errorf("unable to send new notification: %+v", err))
        }
    }
    return nil
}

// function to retrieve analysis timewindow based on current
// timestamp and collection interface specified in connector
func(analyser *TimeseriesAnalyser) GetAnalysisWindow() (time.Time, time.Time) {
    now := time.Now().Round(time.Minute * 1)
    start := now.Add(time.Minute * time.Duration(-analyser.AnalysisIntervalMinutes))
    return start, now
}

// function used to analyse timeseries data
func(analyser *TimeseriesAnalyser) Analyse() error {
    start, end := analyser.GetAnalysisWindow()
    // retrieve list of current businesses from API
    businesses, err := analyser.GetCurrentBusinesses(analyser.TRFAPIConfig)
    if err != nil {
        log.Error(fmt.Errorf("unable to retreive businesses from API: %+v", err))
        return err
    }

    // iterate over businesses and analyse timeseries data for each
    for _, business := range(businesses) {
        if err := analyser.AnalyseBusinessData(business, start, end); err != nil {
            log.Error(fmt.Errorf("unable to analyse business data for %s: %+v", business.BusinessName, err))
            continue
        }
    }
    return nil
}

// function used to start new instance of timeseries analysis engine
func(analyser *TimeseriesAnalyser) Run() {
    log.Info(fmt.Sprintf("starting new timeseries analyzer with analysis interval %d",
        analyser.AnalysisIntervalMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(analyser.AnalysisIntervalMinutes) * time.Minute)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new analysis job...")
                start := time.Now()
                if err := analyser.Analyse(); err != nil {
                    log.Error(fmt.Errorf("unable to analyse data: %+v", err))
                }
                elapsed := time.Now().Sub(start)
                log.Info(fmt.Sprintf("finished analysis job. took %fs to process", elapsed.Seconds()))
            case <- quitChan:
                // stop ticker and add to waitgroup
                ticker.Stop()
                wg.Done()
                return
            }
        }
    }()

    wg.Wait()
    log.Info("stopping timeseries analyser...")
}


