package mail_relay

import (
    "fmt"
    "context"
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
)

var (
    // generate new event channel to process requests
    eventChannel = make(chan MailRelayRequest)

    // define custom errors
    ErrInvalidZipCodeResponse = errors.New("Received invalid response from zipcode API")
    ErrInvalidAPIResponse     = errors.New("Received invalid API response")
    ErrZipCodeNotFound        = errors.New("Cannot find zipcode entry in API")
    ErrUnauthorized           = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse    = errors.New("Received invalid JSON response from zipcode API")
    ErrRequestLimitReached    = errors.New("Reached request limit on API")
)

type MailRelay struct {
    PostgresURL string

    // define connection settings for mail chimp
    MailChimpAPIUrl string
    MailChimpAPIKey string

    // define settings for zip data API
    ZipDataAPIUrl string

    Engine *gin.Engine
}

func New(apiUrl, apiKey, postgresUrl, dataApiUrl string) *MailRelay {
    router := gin.Default()

    // define health check route
    router.GET("/relay/health", healthCheckHandler)
    router.GET("/relay/history", PostgresSessionMiddleware(postgresUrl), relayHistoryHandler)

    // define route to relay mail service
    router.POST("/relay/mail-chimp", PostgresSessionMiddleware(postgresUrl), mailRelayHandler)
    return &MailRelay{
        PostgresURL: postgresUrl,
        MailChimpAPIKey: apiKey,
        MailChimpAPIUrl: apiUrl,
        ZipDataAPIUrl: dataApiUrl,
        Engine: router,
    }
}

func(relay *MailRelay) Run(listenAddress string, listenPort int) {
    // start new go routine to process events
    go relay.ProcessEvents()
    // start gin gonic API to serve requests
    connectionString := fmt.Sprintf("%s:%d", listenAddress, listenPort)
    relay.Engine.Run(connectionString)
}

func(relay *MailRelay) ProcessEvents() {
    // define function used to process mail chimp events async
    log.Info("starting new event processor...")

    for e := range(eventChannel) {
        // defer function used to handle paniced go routine
        defer func() {
            if r := recover(); r != nil {
                log.Warn(fmt.Sprintf("recovered paniced go routine %+v", r))
                // restart go routine
                go relay.ProcessEvents()
            }
        }()

        log.Info(fmt.Sprintf("processing new event %+v", e))
        // create new persistence instance and connect to postgres
        db := NewPersistence(relay.PostgresURL)
        conn, err := db.Connect()
        if err != nil {
            log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
            return
        }

        // get data for zip code (economic region) from API
        zipData, err := relay.GetZipCodeData(e.ZipCode)
        if err != nil {
            log.Error(fmt.Errorf("unable to get zipcode data: %+v", err))
            db.UpdateMailEntry(e.EntryId, "failed", false)
            return
        }
        // relay mail request to Mail Chimp server
        if err := relay.TriggerMailChimp(e, zipData.Data); err != nil {
            log.Error(fmt.Errorf("unable to relay mail request: %+v", err))
            db.UpdateMailEntry(e.EntryId, "failed", false)
            return
        }
        // update relay job in database
        db.UpdateMailEntry(e.EntryId, "completed", true)
        conn.Close(context.Background())
    }
}

func(relay *MailRelay) TriggerMailChimp(request MailRelayRequest, data ZipCodeData) error {
    // function used to relay sign up request to mail chimp server
    log.Info(fmt.Sprintf("relaying request %+v", request))
    return nil
}

func(relay *MailRelay) GetZipCodeData(zipcode string) (ZipCodeDataResponse, error) {
    log.Info(fmt.Sprintf("requesting zip code data for code '%s'", zipcode))

    // createnew HTTP instance and set request headers
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/zipcode/%s", relay.ZipDataAPIUrl, zipcode), nil)
    req.Header.Set("Content-Type", "application/json")

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return ZipCodeDataResponse{}, err
    }
    defer resp.Body.Close()

    // handle response based on status code
    switch resp.StatusCode {
    case 200:
        log.Info(fmt.Sprintf("successfully retrieved zip code data for '%s'", zipcode))
        // parse response body and convert into struct
        results, err := parseZipCodeResponse(resp.Body)
        if err != nil {
            log.Error(fmt.Sprintf("unable to parse JSON response: %+v", err))
            return ZipCodeDataResponse{}, ErrInvalidAPIResponse
        }
        log.Info(fmt.Sprintf("zipcode API returned response %+v", results))
        return results, nil
    case 401:
        log.Error(fmt.Sprintf("received unauthorized response from Zip Data API"))
        return ZipCodeDataResponse{}, ErrUnauthorized
    case 404:
        log.Error(fmt.Sprintf("cannot find API results for zipcode %s", zipcode))
        return ZipCodeDataResponse{}, ErrZipCodeNotFound
    case 429:
        log.Error("reached request limit on API")
        return ZipCodeDataResponse{}, ErrRequestLimitReached
    default:
        log.Error(fmt.Errorf("received invalid response from Zip Data API with code %d", resp.StatusCode))
        return ZipCodeDataResponse{}, ErrInvalidAPIResponse
    }
}
