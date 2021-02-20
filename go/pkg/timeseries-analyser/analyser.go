package timeseries_analyser

import (
	"fmt"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/notifications"
)


type TimeseriesAnalyser struct {
	DataAPIUrl  string
	Notifications notifications.NotificationEngine
}

func NewAnalyser(dataApiUrl string,
	notifications notifications.NotificationEngine) *TimeseriesAnalyser {
	return &TimeseriesAnalyser{
		DataAPIUrl: dataApiUrl,
		Notifications: notifications,
	}
}

// function to retrieve timeseries data from API
func(analyser *TimeseriesAnalyser) GetTimeseriesData(businessId uuid.UUID) {
	log.Debug(fmt.Sprintf("retrieving timeseries data for business %s", businessId))
}
