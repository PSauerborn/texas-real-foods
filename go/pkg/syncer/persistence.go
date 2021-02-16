package syncer

import (
    "fmt"
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v4"
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

// function used to retrieve data for a given business with business ID
func(db *Persistence) GetDataByBusinessId(businessId uuid.UUID) (
    []connectors.BusinessUpdate, error) {
    log.Debug(fmt.Sprintf("retrieving data for business '%s'", businessId))

    data := []connectors.BusinessUpdate{}
    query := `SELECT source,website_live,phone FROM asset_data
    WHERE business_id=$1`

    rows, err := db.Session.Query(context.Background(), query, businessId)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return data, nil
        default:
            return data, err
        }
    }

    for rows.Next() {
        // read variables into local scope
        var (source string; websiteLive bool; phones []string;)
        if err := rows.Scan(&source, &websiteLive, &phones); err != nil {
            log.Warn(fmt.Errorf("unable to read data into local variables: %+v", err))
            continue
        }
        // append new entry to results slice
        data = append(data, connectors.BusinessUpdate{
            Meta: connectors.BusinessMetadata{
                BusinessId: businessId,
            },
            Data: connectors.BusinessData{
                BusinessPhones: phones,
                WebsiteLive: websiteLive,
                Source: source,
            },
        })
    }
    return data, nil
}