package main

import (
    "fmt"
    "strconv"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/parked-domain"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "check_interval_minutes": "1",
            "trf_api_host": "0.0.0.0",
            "trf_api_port": "10999",
            "notify_api_host": "0.0.0.0",
            "notify_api_port": "10756",
            "log_level": "INFO",
        },
    )
)

// function used to generate new notifications API
// config from environment variables
func getNotifyAPIConfig() utils.APIDependencyConfig {
    // get configuration for downstream API dependencies and convert to integer
    apiPortString := cfg.Get("notify_api_port")
    apiPort, err := strconv.Atoi(apiPortString)
    if err != nil {
        panic(fmt.Sprintf("received invalid api port for notify API '%s'", apiPortString))
    }
    return utils.APIDependencyConfig{
        Host: cfg.Get("notify_api_host"),
        Port: &apiPort,
        Protocol: "http",
    }
}

// function used to generate new texas real foods API
// config from environment variables
func getTexasRealFoodsAPIConfig() utils.APIDependencyConfig {
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
    cfg.ConfigureLogging()

    intervalString := cfg.Get("check_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid analysis interval '%s'", intervalString))
    }

    checker := parked_domain.NewDomainChecker(getTexasRealFoodsAPIConfig(),
        getNotifyAPIConfig(), interval)
    checker.Run()

}