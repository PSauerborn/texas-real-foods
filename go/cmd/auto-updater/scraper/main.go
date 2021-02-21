package main

import (
    "fmt"
    "strconv"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors/web"
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
            "trf_api_host": "0.0.0.0",
            "trf_api_port": "10999",
            "utils_api_host": "0.0.0.0",
            "utils_api_port": "10999",
        },
    )
)

func getUtilsAPIConfig() utils.APIDependencyConfig {
    // get configuration for downstream API dependencies and convert to integer
    apiPortString := cfg.Get("utils_api_port")
    apiPort, err := strconv.Atoi(apiPortString)
    if err != nil {
        panic(fmt.Sprintf("received invalid api port for utils API '%s'", apiPortString))
    }
    return utils.APIDependencyConfig{
        Host: cfg.Get("utils_api_host"),
        Port: &apiPort,
        Protocol: "http",
    }
}

func getTRFAPIConfig() utils.APIDependencyConfig {
    // get configuration for downstream API dependencies and convert to integer
    apiPortString := cfg.Get("trf_api_port")
    apiPort, err := strconv.Atoi(apiPortString)
    if err != nil {
        panic(fmt.Sprintf("received invalid api port for trf API '%s'", apiPortString))
    }
    return utils.APIDependencyConfig{
        Host: cfg.Get("trf_api_host"),
        Port: &apiPort,
        Protocol: "http",
    }
}

func main() {
    log.SetLevel(log.DebugLevel)

    // generate new web connector and instance of notification engine
    connector := connectors.NewWebConnector(getUtilsAPIConfig())
    intervalString := cfg.Get("collection_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid collection interval '%s'", intervalString))
    }

    // create new updater with data connector and run
    updater.New(connector, interval, cfg.Get("postgres_url"),
        getTRFAPIConfig()).Run()
}