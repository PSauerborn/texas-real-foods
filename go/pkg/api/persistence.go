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
    query = `INSERT INTO asset_metadata(business_id,business_name,metadata,uri) VALUES($1,$2,$3,$4)`
    _, err = db.Session.Exec(context.Background(), query, businessId,
        request.BusinessName, request.Metadata, request.BusinessURI)
    if err != nil {
        return err
    }

    return nil
}

// function used to retrieve assets from the database
func(db *Persistence) GetAssets() ([]BusinessInfo, error) {
    log.Debug("retrieving assets")
    results := []BusinessInfo{}

    query := `SELECT asset_metadata.business_id, asset_metadata.business_name, asset_metadata.added, asset_metadata.metadata,
        asset_metadata.uri, asset_data.phone, asset_data.website_live, asset_metadata.last_update FROM asset_metadata
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
        var (businessId uuid.UUID; added, lastUpdate time.Time; meta map[string]interface{})
        // handle errors from scanning values into variables
        if err := rows.Scan(&businessId, &businessName, &added, &meta,
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
            Metadata: meta,
        })
    }
    return results, nil
}

func(db *Persistence) GetAssetById(assetId uuid.UUID) (BusinessInfo, error) {
    log.Debug(fmt.Sprintf("retrieving asset with ID %s", assetId))

    query := `SELECT asset_metadata.business_name, asset_metadata.added,
        asset_metadata.uri, asset_data.phone, asset_data.website_live, asset_metadata.last_update, asset_metadata.metadata,
        FROM asset_metadata INNER JOIN asset_data ON asset_metadata.business_id = asset_data.business_id WHERE asset_metadata.asset_id=$1
    `
    var (businessName, businessUri string; added, lastUpdated time.Time)
    var (businessPhones []string; websiteLive bool; meta map[string]interface{})
    // retrieve asset from database
    err := db.Session.QueryRow(context.Background(), query, assetId.String()).Scan(
        &businessName, &added, &businessUri, &businessPhones, &websiteLive,
        &lastUpdated, &meta)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return BusinessInfo{}, ErrAssetNotFound
        default:
            return BusinessInfo{}, err
        }
    }

    info := BusinessInfo{
        BusinessName: businessName,
        BusinessId: assetId,
        BusinessURI: businessUri,
        WebsiteLive: websiteLive,
        BusinessPhones: businessPhones,
        LastUpdate: lastUpdated,
        Added: added,
        Metadata: meta,
    }
    return  info, nil
}

func(db *Persistence) UpdateAssetURI(uri string, assetId uuid.UUID) error {
    log.Debug(fmt.Sprintf("updating asset URI for asset %s", assetId))
    query := `UPDATE asset_metadata SET uri=$1 WHERE business_id=$2`
    _, err := db.Session.Exec(context.Background(), query, uri, assetId)
    return err
}

func(db *Persistence) UpdateAssetMetadata(meta map[string]interface{}, assetId uuid.UUID) error {
    log.Debug(fmt.Sprintf("updating asset URI for asset %s", assetId))
    query := `UPDATE asset_metadata SET metadata=$1 WHERE business_id=$2`
    // update database with new metadata information
    _, err := db.Session.Exec(context.Background(), query, meta, assetId)
    return err
}