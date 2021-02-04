package main

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/utils"
	relay "texas_real_foods/pkg/mail-relay"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_port": "10785",
			"listen_address": "0.0.0.0",
			"zipdata_api_url": "http://localhost:10847",
			"mail_chimp_api_url": "",
			"mail_chimp_api_key": "",
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
	mailServer := relay.New(cfg.Get("mail_chimp_api_url"), cfg.Get("mail_chimp_api_key"),
		cfg.Get("postgres_url"), cfg.Get("zipdata_api_url"))
	mailServer.Run(cfg.Get("listen_address"), listenPort)
}