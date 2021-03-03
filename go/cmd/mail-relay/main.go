package main

import (
    "fmt"
    "strconv"

    "texas_real_foods/pkg/utils"
    relay "texas_real_foods/pkg/mail-relay"
)

var (
    // create map to house environment variables
    cfg = utils.NewConfigMapWithValues(
        map[string]string{
            "listen_port": "10785",
            "listen_address": "0.0.0.0",
            "utils_api_port": "10847",
            "utils_api_host": "0.0.0.0",
            "mail_chimp_api_url": "https://us7.api.mailchimp.com/3.0/",
            "mail_chimp_api_key": "",
            "mail_chimp_api_list_id": "",
            "postgres_url": "postgres://postgres:postgres-dev@192.168.99.100:5432",
        },
    )
)

func getUtilsAPIConfig() utils.APIDependencyConfig {
    // get configuration for downstream API dependencies and convert to integer
    apiPortString := cfg.Get("utils_api_port")
    apiPort, err := strconv.Atoi(apiPortString)
    if err != nil {
        panic(fmt.Sprintf("received invalid api port for trf API '%s'", apiPortString))
    }
    return utils.APIDependencyConfig{
        Host: cfg.Get("utils_api_host"),
        Port: &apiPort,
        Protocol: "http",
    }
}

func getMailChimpConfig() relay.MailChimpConfig {
    // get configuration for downstream API dependencies and convert to integer
    return relay.MailChimpConfig{
        APIKey: cfg.Get("mail_chimp_api_key"),
        APIUrl: cfg.Get("mail_chimp_api_url"),
        ListID: cfg.Get("mail_chimp_api_list_id"),
    }
}

func main() {
    cfg.ConfigureLogging()
    // get listen port from environment variables and start new server
    listenPort, err := strconv.Atoi(cfg.Get("listen_port"))
    if err != nil {
        panic(fmt.Sprintf("invalid listen port '%s'", cfg.Get("listen_port")))
    }

    // generate new postgres persistence
    persistence := relay.NewPersistence(cfg.Get("postgres_url"))
    conn, err := persistence.Connect()
    if err != nil {
        panic(fmt.Errorf("unable to connect to postgres server: %+v", err))
    }
    defer conn.Close()

    // generate new mail server and run
    mailServer := relay.NewMailRelay(getMailChimpConfig(), getUtilsAPIConfig(),
        persistence)
    mailServer.Run(fmt.Sprintf(":%d", listenPort))
}