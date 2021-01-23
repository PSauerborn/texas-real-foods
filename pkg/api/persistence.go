package api

import (
    "fmt"
    "time"
    "context"

    "github.com/jackc/pgx/v4"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)


type Persistence struct{
    DatabaseURL string
    Session     *pgx.Conn
}

func NewPersistence(url string) *Persistence {
    return &Persistence{
        DatabaseURL: url,
    }
}

// function to connect persistence to postgres server
// note that the connection is returned and should be
// closed with a defer conn.Close(context) statement
func(db *Persistence) Connect() (*pgx.Conn, error) {
    log.Debug(fmt.Sprintf("creating new database connection"))
    // connect to postgres server and set session in persistence
    conn, err := pgx.Connect(context.Background(), db.DatabaseURL)
    if err != nil {
        log.Error(fmt.Errorf("error connecting to postgres service: %+v", err))
        return nil, err
    }
    db.Session = conn
    return conn, err
}

// function to insert new asset into database
func(db *Persistence) CreateAsset(request NewAssetRequest) error {
    log.Debug(fmt.Sprintf("creating new asset %+v", request))

    var (query string; err error)
    businessId := uuid.New()

    // insert data into metadata table
    query = `INSERT INTO asset_metadata(business_id,business_name) VALUES($1,$2)`
    _, err = db.Session.Exec(context.Background(), query, businessId, request.BusinessName)
    if err != nil {
        return err
    }

    // insert data into asset data table
    query = `INSERT INTO asset_data(business_id,uri) VALUES($1,$2)`
    _, err = db.Session.Exec(context.Background(), query, businessId, request.BusinessURI)
    if err != nil {
        return err
    }

    return nil
}

// function used to retrieve assets from the database
func(db *Persistence) GetAssets() ([]BusinessInfo, error) {
    log.Debug("retrieving assets")
    results := []BusinessInfo{}

    query := `SELECT asset_metadata.business_id, asset_metadata.business_name, asset_metadata.added,
        asset_data.uri, asset_data.phone, asset_data.website_live, asset_metadata.last_update FROM asset_metadata
        INNER JOIN asset_data ON asset_metadata.business_id = asset_data.business_id
    `
    // retrieve assets from database
    rows, err := db.Session.Query(context.Background(), query)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return results, nil
        default:
            return results, err
        }
    }

    for rows.Next() {
        var (businessName, businessUri string; businessPhones []string; websiteLive bool)
        var (businessId uuid.UUID; added, lastUpdate time.Time)
        // handle errors from scanning values into variables
        if err := rows.Scan(&businessId, &businessName, &added,
            &businessUri, &businessPhones, &websiteLive, &lastUpdate); err != nil {
            log.Error(fmt.Errorf("unable to scan database row: %+v", err))
            continue
        }
        // append new business info to results
        results = append(results, BusinessInfo{
            BusinessId: businessId,
            BusinessName: businessName,
            BusinessURI: businessUri,
            BusinessPhones: businessPhones,
            WebsiteLive: websiteLive,
            Added: added,
            LastUpdate: lastUpdate,
        })
    }
    return results, nil
}