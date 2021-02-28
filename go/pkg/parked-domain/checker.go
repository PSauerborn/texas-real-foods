package parked_domain

import (
    "fmt"
    "time"
    "sync"
    "net/http"
    "io/ioutil"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/notifications"
    apis "texas_real_foods/pkg/utils/api_accessors"
)


type ParkedDomainChecker struct {
    TexasRealFoodsAPIConfig utils.APIDependencyConfig
    NotificationsAPIConfig utils.APIDependencyConfig
    CheckIntervalMinutes int
}

func NewDomainChecker(texasRealFoodsAPIConfig utils.APIDependencyConfig,
    notifyAPIConfig utils.APIDependencyConfig, interval int) *ParkedDomainChecker {
    return &ParkedDomainChecker{
        TexasRealFoodsAPIConfig: texasRealFoodsAPIConfig,
        NotificationsAPIConfig: notifyAPIConfig,
        CheckIntervalMinutes: interval,
    }
}

// function used to retrieve business metadata for all stored
// businesses
func(checker *ParkedDomainChecker) GetCurrentBusinesses() ([]connectors.BusinessMetadata, error) {
    // establish new connection to postgres persistence
    access := apis.NewTRFApiAccessorFromConfig(checker.TexasRealFoodsAPIConfig)
    payload, err := access.GetBusinesses()
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve businesses from API: %+v", err))
        return payload.Data, err
    }
    return payload.Data, nil
}

// function used to check for parked domains
func(checker *ParkedDomainChecker) IsDomainParked(business connectors.BusinessMetadata) (bool, error) {
    // generate new HTTP request
    request, err := http.NewRequest("GET", business.BusinessURI, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP request for %s: %+v",
            business.BusinessName, err))
        return false, err
    }
    // generate new client instance and retrieve site data
    client := http.Client{}
    response, err := client.Do(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return false, err
    }
    defer response.Body.Close()

    switch response.StatusCode {
    case 200:
        log.Debug("successfully retrieved site data. checking for parked domain...")
        data, err := ioutil.ReadAll(response.Body)
        if err != nil {
            log.Error(fmt.Sprintf("unable to read response body: %+v", err))
            return false, err
        }
        // iterate over conditions and check for expired domains
        for _, condition := range(ParkedDomainConditions) {
            if condition(string(data)) {
                log.Info(fmt.Sprintf("parked domain condition met for %s at %s", business.BusinessName,
                    business.BusinessURI))
                return true, nil
            }
        }
    default:
        log.Error(fmt.Errorf("unable to get data for URI %s: received response code %d",
            business.BusinessURI, response.StatusCode))
        return false, err
    }
    return false, nil
}

// function used to create new notification with API for
// any businesses that contain parked domains
func(checker *ParkedDomainChecker) SendParkedNotification(business connectors.BusinessMetadata) error {
    notificationString := fmt.Sprintf("Found parked domain for business %s at URI %s",
        business.BusinessName, business.BusinessURI)
    // generate new notification
    notification := notifications.ChangeNotification{
        BusinessId: business.BusinessId,
        BusinessName: business.BusinessName,
        EventTimestamp: time.Now(),
        Notification: notificationString,
        NotificationHash: generateNotificationHash(business.BusinessId),
        Metadata: map[string]interface{}{
            "source": "parked-domain-checker",
        },
    }

    // create new API accessor and send notification to API
    accessor := apis.NewNotificationsApiAccessorFromConfig(checker.NotificationsAPIConfig)
    // send notification for business change
    if _, err := accessor.CreateNotification(notification); err != nil {
        log.Error(fmt.Errorf("unable to send new notification: %+v", err))
        return err
    }
    return nil
}

// function used to run new instance of parked domain checker
func(checker *ParkedDomainChecker) Run() {
    log.Info(fmt.Sprintf("starting new parked domain checker with interval %d...",
        checker.CheckIntervalMinutes))
    // generate ticker and channel for messages
    ticker := time.NewTicker(time.Duration(checker.CheckIntervalMinutes) * time.Minute)
    quitChan := make(chan bool)

    var wg sync.WaitGroup
    // add to waitgroup to prevent go routine from closing
    wg.Add(1)

    go func() {
        for {
            select {
            case <- ticker.C:
                log.Info("starting new domain parked check job...")
                start := time.Now()
                // get all current businesses from TRF API
                businesses, err := checker.GetCurrentBusinesses()
                if err != nil {
                    log.Error(fmt.Errorf("unable to retrieve businesses from API: %+v", err))
                    return
                }

                // iterate over businesses and checked for parked domains
                for _, business := range(businesses) {
                    parked, err := checker.IsDomainParked(business)
                    if err != nil {
                        log.Error(fmt.Errorf("unable to check business %s at URI %s for parked domain: %+v",
                            business.BusinessName, business.BusinessURI, err))
                        continue
                    } else if parked {
                        if err := checker.SendParkedNotification(business); err != nil {
                            log.Warn(fmt.Sprintf("unable to send notification to API: %+v", err))
                            continue
                        }
                    }
                }

                // log total time elapsed to process job
                elapsed := time.Now().Sub(start)
                log.Info(fmt.Sprintf("finished clearing job. took %fs to process", elapsed.Seconds()))
            case <- quitChan:
                // stop ticker and add to waitgroup
                ticker.Stop()
                wg.Done()
                return
            }
        }
    }()
    wg.Wait()
    log.Info("stopping parked-domain checker...")
}