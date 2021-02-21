package main

import (
	"fmt"
	"strconv"

	"texas_real_foods/pkg/timeseries-analyser"
	"texas_real_foods/pkg/notifications"
	"texas_real_foods/pkg/utils"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
            "analysis_interval_minutes": "1",
            "trf_api_host": "0.0.0.0",
            "trf_api_port": "10999",
        },
    )
)

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
    }
}

func main() {
	intervalString := cfg.Get("analysis_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid analysis interval '%s'", intervalString))
	}

	notify := notifications.NewDefaultNotificationEngine(cfg.Get("postgres_url"))
	timeseries_analyser.NewAnalyser(getTRFAPIConfig(), notify, interval).Run()
}