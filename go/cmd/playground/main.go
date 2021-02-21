package main

import (
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/utils/api_accessors"
)

func main() {
	log.SetLevel(log.DebugLevel)

	access := utils.NewTRFApiAccessor("trf.project-gateway.app/api", "https", nil)
	log.Info(access.GetBusinesses())
}