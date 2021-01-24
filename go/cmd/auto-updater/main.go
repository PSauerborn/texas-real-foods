package main

import (
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors/web"
    "texas_real_foods/pkg/notifications"
    "texas_real_foods/pkg/utils"
    updater "texas_real_foods/pkg/auto-updater"
)

var (
    // create map to house environment variables
    environConfig = utils.NewConfigMapWithValues(
        map[string]string{
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
            "phone_validation_api_host": "http://localhost:10847",
        },
    )
)

func main() {
    log.SetLevel(log.DebugLevel)

    // generate new web connector and instance of notification engine
    connector := connectors.NewWebConnector(environConfig.Get("phone_validation_api_host"))
    notify := notifications.NewDefaultNotificationEngine(environConfig.Get("postgres_url"))

    // create new updater with data connector and run
    updater.New(connector, 10, notify, environConfig.Get("postgres_url")).Run()
}