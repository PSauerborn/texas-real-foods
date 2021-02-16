package authenticator

import (
    "fmt"
    "context"

    "github.com/jackc/pgx/v4"
    log "github.com/sirupsen/logrus"

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

// function to retrieve all API access keys from database
func(db *Persistence) GetAPIKeys() ([]string, error) {
    log.Debug("retrieving API keys")

    results := []string{}

    query := `SELECT key FROM access_keys`
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
        var key string
        if err := rows.Scan(&key); err != nil {
            log.Warn(fmt.Errorf("unable to read key into local variable: %+v", err))
            continue
        }
        results = append(results, key)
    }
    return results, nil
}

// function used to check if a given API key is valid
func(db *Persistence) IsValidApiKey(key string) (bool, error) {
    log.Debug(fmt.Sprintf("validating API key %s", key))

    var result string
    query := `SELECT key FROM access_keys WHERE key=$1`
    row := db.Session.QueryRow(context.Background(), query, key)
    if err := row.Scan(&result); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return false, nil
        default:
            return false, err
        }
    }
    return true, nil
}
