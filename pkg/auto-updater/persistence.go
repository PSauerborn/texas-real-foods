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

// function used to retrieve list of current assets
func(db *Persistence) GetAllAssets() ([]connectors.BusinessInfo, error) {
	log.Debug("retrieving list of current assets")

	results := []connectors.BusinessInfo{}

	query := `SELECT asset_metadata.business_id, asset_metadata.business_name,
		asset_data.uri, asset_data.phone, asset_data.website_live FROM asset_metadata
		INNER JOIN asset_data ON asset_metadata.business_id = asset_data.business_id
	`
	rows, err := db.Session.Query(context.Background(), query)
	if err != nil {
		switch err {
		case pgx.ErrNoRows:
			return results, nil
		default:
			return results, err
		}
	}

	// iterate over database rows and add to results
	for rows.Next() {
		var (businessName, businessUri string; businessPhones []string)
		var (businessId uuid.UUID; siteLive bool)

		// scan rows into variables
		if err := rows.Scan(&businessId, &businessName, &businessUri,
			&businessPhones, &siteLive); err != nil {
			log.Error(fmt.Errorf("unable to scan database row: %+v", err))
			continue
		}
		// append results
		results = append(results, connectors.BusinessInfo{
			BusinessId: businessId,
			BusinessName: businessName,
			BusinessPhones: businessPhones,
			BusinessURI: businessUri,
			WebsiteLive: siteLive,
		})
	}
	return results, nil
}

func(db *Persistence) UpdateAsset(asset connectors.BusinessInfo) error {
	log.Debug(fmt.Sprintf("updating asset %+v", asset))

	var (query string; err error)

	// update metadata table with update timestamp
	query = `UPDATE asset_metadata SET last_update=$1 WHERE business_id=$2`
	_, err = db.Session.Exec(context.Background(), query, time.Now(), asset.BusinessId)
	if err != nil {
		return err
	}
	// update data table with new asset values
	query = `UPDATE asset_data SET phone=$1 WHERE business_id=$2`
	_, err = db.Session.Exec(context.Background(), query, asset.BusinessPhones, asset.BusinessId)
	if err != nil {
		return err
	}
	return err
}