package auto_updater

import (
    "fmt"
    "time"
    "context"

    "github.com/jackc/pgx/v4"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/connectors"
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

// function to retrieve all business metadata from the postgres
// database
func(db *Persistence) GetAllMetadata() ([]connectors.BusinessMetadata, error) {
    log.Debug("retrieving business metadata")

    results := []connectors.BusinessMetadata{}
    query := `SELECT business_id,business_name,metadata,uri FROM asset_metadata`
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
        var (businessId uuid.UUID; businessName, businessUri string)
        var meta map[string]interface{}

        if err := rows.Scan(&businessId, &businessName, &meta,
            &businessUri); err != nil {
            log.Warn(fmt.Errorf("unable to scan values into local variables: %+v", err))
            continue
        }
        results = append(results, connectors.BusinessMetadata{
            BusinessId: businessId,
            BusinessName: businessName,
            BusinessURI: businessUri,
            Metadata: meta,
        })
    }
    return results, nil
}

// function to update business data in the database. note that
// updates are done as inserts i.e. existing values are overwritten
func(db *Persistence) UpdateBusinessData(update connectors.BusinessUpdate) error {
    log.Debug(fmt.Sprintf("updating business %+v", update))

    var query string

    // execute query to insert new data arguments
    query = `INSERT INTO asset_data(business_id,phone,website_live,source)
    VALUES($1,$2,$3,$4) ON CONFLICT (business_id,source) DO UPDATE
    SET phone=$2, website_live=$3`
    _, err := db.Session.Exec(context.Background(), query, update.Meta.BusinessId,
    update.Data.BusinessPhones, update.Data.WebsiteLive, update.Data.Source)
    if err != nil {
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