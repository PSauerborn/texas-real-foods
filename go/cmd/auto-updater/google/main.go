package main

import (
    "fmt"
    "strconv"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors/google"
    "texas_real_foods/pkg/utils"
    updater "texas_real_foods/pkg/auto-updater"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
            "collection_interval_minutes": "1",
            "google_base_api": "https://maps.googleapis.com/maps/api/place/details/json",
            "google_api_key": "",
            "base_api_url": "http://0.0.0.0:10999/texas-real-foods",
        },
    )
)

func main() {
    log.SetLevel(log.DebugLevel)

    // generate new web connector and instance of notification engine
    connector := connectors.NewGoogleAPIConnector(cfg.Get("google_base_api"),
        cfg.Get("google_api_key"))

    intervalString := cfg.Get("collection_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid collection interval '%s'", intervalString))
    }
    // create new updater with data connector and run
    updater.New(connector, interval, cfg.Get("postgres_url"),
        cfg.Get("base_api_url")).Run()
}