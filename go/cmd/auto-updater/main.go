package main

import (
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors/web"
    updater "texas_real_foods/pkg/auto-updater"
)

func main() {
    log.SetLevel(log.DebugLevel)

    // generate new web connector and run new instance of updater
    connector := connectors.NewWebConnector()
    updater.New(connector, 10).Run()
}