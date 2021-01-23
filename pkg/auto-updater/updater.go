package auto_updater

import (
	"fmt"
	"time"
	"sync"
	"context"

	log "github.com/sirupsen/logrus"

	"texas_real_foods/pkg/connectors"
)

var (
	postgresURL = "postgres://postgres:postgres-dev@192.168.99.100:5432/postgres"
)

// function used to create new auto updater
func New(connector connectors.AutoUpdateDataConnector,
	collectionPeriod int) *AutoUpdater {
	return &AutoUpdater{
		DataConnector: connector,
		CollectionPeriodMinutes: collectionPeriod,
	}
}

type AutoUpdater struct{
	DataConnector connectors.AutoUpdateDataConnector
	CollectionPeriodMinutes int
}

// function used to process asset information
func(updater *AutoUpdater) GetCurrentAssets() ([]connectors.BusinessInfo, error) {
	// establish new connection to postgres persistence
	db := NewPersistence(postgresURL)
	conn, err := db.Connect()
	if err != nil {
		log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
		return []connectors.BusinessInfo{}, err
	}
	defer conn.Close(context.Background())
	// get all assets from database and return
	return db.GetAllAssets()
}

// function used to process asset information
func(updater *AutoUpdater) ProcessAssets(assets []connectors.BusinessInfo) error {
	// establish new connection to postgres persistence
	db := NewPersistence(postgresURL)
	conn, err := db.Connect()
	if err != nil {
		log.Error(fmt.Errorf("unable to retrieve assets from postgres: %+v", err))
		return err
	}
	defer conn.Close(context.Background())

	// iterate over assets and update
	for _, asset := range(assets) {
		if err := db.UpdateAsset(asset); err != nil {
			log.Error(fmt.Errorf("unable to update asset '%s': %+v", asset.BusinessName, err))
		}
	}
	return nil
}

// function used to start updater. the auto updater
func(updater *AutoUpdater) Run() {
	log.Info(fmt.Sprintf("starting new asset auto-updater with collection interval %d", updater.CollectionPeriodMinutes))
	// generate ticker and channel for messages
	ticker := time.NewTicker(time.Duration(updater.CollectionPeriodMinutes) * time.Second)
	quitChan := make(chan bool)

	var wg sync.WaitGroup
	// add to waitgroup to prevent go routine from closing
	wg.Add(1)

	go func() {
		for {
			select {
			case <- ticker.C:
				log.Info("starting new collection job...")
				// retrieve current list of assets
				currentAssets, err := updater.GetCurrentAssets()
				if err != nil {
					log.Error(fmt.Errorf("unable to retrieve existing assets: %+v", err))
					continue
				}

				// retrieve updated asset list from connector
				assets, err := updater.DataConnector.CollectData(currentAssets)
				if err != nil {
					log.Error(fmt.Errorf("unable to retrieve asset data: %+v", err))
					continue
				}

				// process collected assets
				if len(assets) > 0 {
					log.Info(fmt.Sprintf("successfully retrieved %d assets. processing...", len(assets)))
					updater.ProcessAssets(assets)
				} else {
					log.Info("no changes in assets detected. sleeping...")
				}

			case <- quitChan:
				// stop ticker and add to waitgroup
				ticker.Stop()
				wg.Done()
				return
			}
		}
	}()

	wg.Wait()
	log.Info("stopping auto-updater...")
}

