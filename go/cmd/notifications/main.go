package main

import (
    "fmt"
    "strconv"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/notifications"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_port": "10756",
            "listen_address": "0.0.0.0",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
        },
    )
)

func main() {
    log.SetLevel(log.DebugLevel)

    // get listen port from environment variables and start new server
    listenPort, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(fmt.Sprintf("invalid listen port '%s'", cfg.Get("listen_port")))
    }
    // create new instance of mail relay with variables and run
    service := notifications.NewNotificationService(cfg.Get("postgres_url"))
    service.Run(fmt.Sprintf("%s:%d", cfg.Get("listen_address"), listenPort))
}