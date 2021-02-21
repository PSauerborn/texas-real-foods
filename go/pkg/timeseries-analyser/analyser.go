package timeseries_analyser

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/notifications"
	"texas_real_foods/pkg/utils"
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


