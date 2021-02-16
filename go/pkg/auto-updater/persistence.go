package auto_updater

import (
    "fmt"
    "time"
    "context"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
    "texas_real_foods/pkg/utils"
)


type Persistence struct{
    *utils.BasePostgresPersistence
}

func NewPersistence(url string) *Persistence {
    // create instance of base persistence
    basePersistence := utils.NewPersistence(url)
    return &Persistence{
        basePersistence,
    }
}

// function to update business data in the database. note that
// updates are done as inserts i.e. existing values are overwritten
func(db *Persistence) UpdateBusinessData(update connectors.BusinessUpdate) error {
    log.Debug(fmt.Sprintf("updating business %+v", update))

    var (query string; err error)
    // execute query to insert new data arguments
    query = `INSERT INTO asset_data(business_id,phone,website_live,source,open)
        VALUES($1,$2,$3,$4,$5) ON CONFLICT (business_id,source) DO UPDATE
        SET phone=$2, website_live=$3, open=$5`
    _, err = db.Session.Exec(context.Background(), query, update.Meta.BusinessId,
        update.Data.BusinessPhones, update.Data.WebsiteLive, update.Data.Source, update.Data.BusinessOpen)
    if err != nil {
        log.Error(fmt.Errorf("unable to insert data into static table: %+v", err))
        return err
    }

    // add entry into time series table
    query = `INSERT INTO asset_data_timeseries(business_id,source,phone,website_live,open)
        VALUES($1,$2,$3,$4,$5)`
    _, err = db.Session.Exec(context.Background(), query, update.Meta.BusinessId,
        update.Data.Source, update.Data.BusinessPhones, update.Data.WebsiteLive, update.Data.BusinessOpen)
    if err != nil {
        log.Error(fmt.Errorf("unable to insert data into timeseries table: %+v", err))
        return err
    }
    // update metadata with last update flag
    query = `UPDATE asset_metadata SET last_update=$1 WHERE business_id=$2`
    _, err = db.Session.Exec(context.Background(), query, time.Now(), update.Meta.BusinessId)
    if err != nil {
        return err
    }
    return nil
}