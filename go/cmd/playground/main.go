package main

import (
	"fmt"
	"time"
	"context"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/utils"
	"texas_real_foods/pkg/notifications"
	apis "texas_real_foods/pkg/utils/api_accessors"
	"texas_real_foods/pkg/connectors"
)

var (
	postgresUrl = "postgres://postgres:postgres-dev@192.168.99.100:5432"
)

type Persistence struct {
	*utils.BasePostgresPersistence
}

func NewPersistence(url string) *Persistence {
    // create instance of base persistence
    basePersistence := utils.NewPersistence(url)
    return &Persistence{
        basePersistence,
    }
}

type TimeseriesEntry struct {
	EventTimestamp time.Time
	connectors.BusinessData
}

func(db *Persistence) GenerateTimeSeriesData(businessId uuid.UUID, source string) {
	log.Debug("generating new timeseries...")
	// define start and end time
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	query := `INSERT INTO asset_data_timeseries(business_id,source,event_timestamp,
		website_live,phone,open) VALUES($1,$2,$3,$4,$5,$6)`

	for {
		// break out of loop if start time has exceeded
		if startTime.After(endTime) {
			_, err := db.Session.Exec(context.Background(), query, businessId, source,
				startTime, true, []string{"0286605553", "1555456789"}, false)
			if err != nil {
				log.Error(fmt.Errorf("unable to insert data entry: %+v", err))
			}
			break
		} else {
			// insert new values into database
			_, err := db.Session.Exec(context.Background(), query, businessId, source,
				startTime, true, []string{"20286605553"}, true)
			if err != nil {
				log.Error(fmt.Errorf("unable to insert data entry: %+v", err))
			}
		}
		// increment start time
		startTime = startTime.Add(5 * time.Minute)
	}
}

func main() {
	log.SetLevel(log.DebugLevel)

	// businessId, _ := uuid.Parse("42e15fa3-c07f-46c8-88ea-b42b38ad352d")
	// // establish new connection to postgres persistence
    // db := NewPersistence(postgresUrl)
    // conn, err := db.Connect()
    // if err != nil {
    //     log.Error(fmt.Errorf("unable to connect to postgres server: %+v", err))
    //     return
    // }
    // defer conn.Close()

	// sources := []string{"yelp-api-connector", "scraper"}
	// for _, source := range(sources) {
	// 	db.GenerateTimeSeriesData(businessId, source)
	// }

	port := 10756
	cfg := utils.APIDependencyConfig{
		Host: "0.0.0.0",
		Port: &port,
		Protocol: "http",
	}

	notification := notifications.ChangeNotification{
		BusinessId: uuid.New(),
		BusinessName: "test-business",
		NotificationHash: "test-hash",
		EventTimestamp: time.Now(),
		Notification: "testing-notification",
	}

	accessor := apis.NewNotificationsApiAccessorFromConfig(cfg)
	log.Info(accessor.CreateNotification(notification))
}