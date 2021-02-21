package main

import (
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/utils/api_accessors"
)

func main() {
	log.SetLevel(log.DebugLevel)

	port := 10999
	access := utils.NewTRFApiAccessor("0.0.0.0", "http", &port)

	businessId, _ := uuid.Parse("9a2cae53-1104-4688-b3d9-53953f23f003")
	now := time.Now()
	log.Info(access.GetTimeseriesData(businessId, now, now.Add(time.Duration(15) * time.Minute)))
}