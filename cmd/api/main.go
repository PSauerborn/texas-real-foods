package main

import (
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/api"
)

func main() {
	log.SetLevel(log.DebugLevel)

	router := api.New()
	router.Run(":10999")
}