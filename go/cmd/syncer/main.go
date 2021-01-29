package main

import (
	"fmt"
	"strconv"

	"texas_real_foods/pkg/syncer"
	"texas_real_foods/pkg/notifications"
	"texas_real_foods/pkg/utils"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
			"postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
			"collection_interval_minutes": "1",
        },
    )
)

func main() {

	intervalString := cfg.Get("collection_interval_minutes")
    // convert given interval from string to integer
    interval, err := strconv.Atoi(intervalString)
    if err != nil {
        panic(fmt.Sprintf("received invalid collection interval '%s'", intervalString))
	}

	notify := notifications.NewDefaultNotificationEngine(cfg.Get("postgres_url"))
	worker := syncer.NewSyncer(cfg.Get("postgres_url"), interval, notify)
	worker.Run()
}