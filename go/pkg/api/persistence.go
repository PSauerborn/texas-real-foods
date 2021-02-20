package api

import (
    "fmt"
    "time"
    "context"

    "github.com/jackc/pgx/v4"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/connectors"
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

// function to insert new business into database
func(db *Persistence) CreateBusiness(request NewBusinessRequest) error {
    log.Debug(fmt.Sprintf("creating new businesses %+v", request))

    var (query string; err error)
    businessId := uuid.New()

    // insert data into metadata table
    query = `INSERT INTO asset_metadata(business_id,business_name,metadata,uri) VALUES($1,$2,$3,$4)`
    _, err = db.Session.Exec(context.Background(), query, businessId,
        request.BusinessName, request.Metadata, request.BusinessURI)
    if err != nil {
        return err
    }

    return nil
}

// function used to retrieve businesses from the database
func(db *Persistence) GetBusinesses() ([]BusinessInfo, error) {
    log.Debug("retrieving businesses")
    results := []BusinessInfo{}

    query := `SELECT asset_metadata.business_id, asset_metadata.business_name,
    asset_metadata.added, asset_metadata.metadata, asset_metadata.uri,
    asset_metadata.last_update FROM asset_metadata`

    // retrieve businesses from database
    rows, err := db.Session.Query(context.Background(), query)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            log.Warn("no businesses found in database. returning empty array")
            return results, nil
        default:
            return results, err
        }
    }

    for rows.Next() {
        var businessName, businessUri string
        var (businessId uuid.UUID; added, lastUpdate time.Time; meta map[string]interface{})
        // handle errors from scanning values into variables
        if err := rows.Scan(&businessId, &businessName, &added, &meta,
            &businessUri, &lastUpdate); err != nil {
            log.Error(fmt.Errorf("unable to scan database row: %+v", err))
            continue
        }
        // append new business info to results
        results = append(results, BusinessInfo{
            BusinessId: businessId,
            BusinessName: businessName,
            BusinessURI: businessUri,
            Added: added,
            LastUpdate: lastUpdate,
            Metadata: meta,
        })
    }
    return results, nil
}

func(db *Persistence) GetBusinessById(businessId uuid.UUID) (BusinessInfo, error) {
    log.Debug(fmt.Sprintf("retrieving businesses with ID %s", businessId))

    query := `SELECT asset_metadata.business_name, asset_metadata.added,
        asset_metadata.uri, asset_metadata.last_update, asset_metadata.metadata
        FROM asset_metadata WHERE business_id=$1`

    var (businessName, businessUri string; added, lastUpdated time.Time)
    var meta map[string]interface{}
    // retrieve business from database
    err := db.Session.QueryRow(context.Background(), query, businessId.String()).Scan(
        &businessName, &added, &businessUri, &lastUpdated, &meta)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return BusinessInfo{}, ErrBusinessNotFound
        default:
            return BusinessInfo{}, err
        }
    }

    info := BusinessInfo{
        BusinessName: businessName,
        BusinessId: businessId,
        BusinessURI: businessUri,
        LastUpdate: lastUpdated,
        Added: added,
        Metadata: meta,
    }
    return  info, nil
}

func(db *Persistence) UpdateBusinessURI(uri string, businessId uuid.UUID) error {
    log.Debug(fmt.Sprintf("updating business URI for business %s", businessId))
    query := `UPDATE asset_metadata SET uri=$1 WHERE business_id=$2`
    _, err := db.Session.Exec(context.Background(), query, uri, businessId)
    return err
}

func(db *Persistence) UpdateBusinessMetadata(meta map[string]interface{}, businessId uuid.UUID) error {
    log.Debug(fmt.Sprintf("updating business URI for business %s", businessId))
    query := `UPDATE asset_metadata SET metadata=$1 WHERE business_id=$2`
    // update database with new metadata information
    _, err := db.Session.Exec(context.Background(), query, meta, businessId)
    return err
}

// function to delete a business with given business id
func(db *Persistence) DeleteBusiness(businessId uuid.UUID) error {
    log.Debug(fmt.Sprintf("deleting business %s", businessId))

    var (err error; query string)

    query = `DELETE FROM asset_metadata WHERE business_id=$1`
    _, err = db.Session.Exec(context.Background(), query, businessId)
    if err != nil {
        return err
    }

    query = `DELETE FROM asset_data WHERE business_id=$1`
    _, err = db.Session.Exec(context.Background(), query, businessId)
    if err != nil {
        return err
    }
    return nil
}

// function to retreive notifications from database
func(db *Persistence) GetNotifications() ([]Notification, error) {
    log.Debug("retrieving notifications from database")

    notifications := []Notification{}

    query := `SELECT notification_id, event_timestamp, notification, hash
    FROM notifications`
    rows, err := db.Session.Query(context.Background(), query)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return notifications, nil
        default:
            return notifications, err
        }
    }

    for rows.Next() {
        var (notificationId uuid.UUID; eventTimestamp time.Time)
        var (notification map[string]interface{}; hashed string)

        if err := rows.Scan(&notificationId, &eventTimestamp,
            &notification, &hashed); err != nil {
            log.Warn(fmt.Errorf("unable to scan row into local variables: %+v", err))
            continue
        }

        notifications = append(notifications, Notification{
            NotificationId: notificationId,
            EventTimestamp: eventTimestamp,
            Notification: notification,
            Hash: hashed,
        })
    }
    return notifications, nil
}

// function to retrieve static data for a given business with business ID
func(db *Persistence) GetStaticBusinessData(businessId uuid.UUID) ([]connectors.BusinessData, error) {
    log.Debug(fmt.Sprintf("retrieving static data for business %s", businessId))

    results := []connectors.BusinessData{}
    query := `SELECT phone,website_live,open,source FROM asset_data
        WHERE business_id=$1`
    rows, err := db.Session.Query(context.Background(), query, businessId)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return results, nil
        default:
            return results, err
        }
    }

    for rows.Next() {
        // scan data into local variables
        var data connectors.BusinessData
        if err := rows.Scan(&data.BusinessPhones, &data.WebsiteLive, &data.BusinessOpen,
            &data.Source); err != nil {
            log.Warn(fmt.Errorf("unable to scan data into local variables: %+v", err))
            continue
        }
        // adddata entry to results array
        results = append(results, data)
    }
    return results, nil
}

// struct to that extends business data struct with event timestamp
type TimeSeriesData struct {
    EventTimestamp time.Time `json:"event_timestamp"`
    connectors.BusinessData
}

// function to retrive timeseries data from database
func(db *Persistence) GetTimeSeriesData(businessId uuid.UUID, start,
    end time.Time) ([]TimeSeriesData, error) {

    log.Debug(fmt.Sprintf("retrieving timeseries business data for business %s", businessId))
    results := []TimeSeriesData{}
    query := `SELECT phone,website_live,open,source,event_timestamp
        FROM asset_data_timeseries WHERE business_id=$1 AND event_timestamp > $2 AND event_timestamp < $3`
    // query rows from postgres database
    rows, err := db.Session.Query(context.Background(), query, businessId, start, end)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return results, nil
        default:
            return results, err
        }
    }

    for rows.Next() {
        // scan data into local variables
        var (data connectors.BusinessData; ts time.Time)
        if err := rows.Scan(&data.BusinessPhones, &data.WebsiteLive, &data.BusinessOpen,
            &data.Source, &ts); err != nil {
            log.Warn(fmt.Errorf("unable to scan data into local variables: %+v", err))
            continue
        }
        // adddata entry to results array
        results = append(results, TimeSeriesData{ts, data})
    }
    return results, nil
}