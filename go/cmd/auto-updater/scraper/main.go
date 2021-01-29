package main

import (
    "fmt"
    "strconv"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors/web"
    "texas_real_foods/pkg/notifications"
    "texas_real_foods/pkg/utils"
    updater "texas_real_foods/pkg/auto-updater"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
            "phone_validation_api_host": "http://localhost:10847",
            "collection_interval_minutes": "1",
        },
    )
)

func main() {
    log.SetLevel(log.DebugLevel)

    // generate new web connector and instance of notification engine
    // connector := connectors.NewWebConnector(environConfig.Get("phone_validation_api_host"))
    connector := connectors.NewWebConnector(cfg.Get("phone_validation_api_host"))
    notify := notifications.NewDefaultNotificationEngine(cfg.Get("postgres_url"))

    intervalString := cfg.Get("collection_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid collection interval '%s'", intervalString))
    }
    // create new updater with data connector and run
    updater.New(connector, interval, notify, cfg.Get("postgres_url")).Run()
}