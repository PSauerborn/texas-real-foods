package mail_relay

import (
    "fmt"
    "context"
    "time"
    "errors"
    "encoding/json"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v4"
    log "github.com/sirupsen/logrus"
)

var (
    ErrInvalidJsonEntry = errors.New("Invalid entry: entries must be JSON format")
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

func(db *Persistence) InsertMailEntry(request MailRelayRequest) (uuid.UUID, error) {
    log.Debug(fmt.Sprintf("inserting new mail entry %+v", request))
    entryId := uuid.New()
    request.EntryId = entryId

    entryJson, err := json.Marshal(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert entry to JSON: %+v", err))
        return entryId, ErrInvalidJsonEntry
    }

    // insert new entry into postgres database
    query := `INSERT INTO mail_relay_entries(entry_id,status,completed,data) VALUES($1,$2,$3,$4)`
    _, err = db.Session.Exec(context.Background(), query, entryId, "in progress", false, entryJson)
    if err != nil {
        log.Error(fmt.Errorf("unable to entry new relay entry: %+v", err))
        return entryId, err
    }
    return entryId, nil
}

func(db *Persistence) UpdateMailEntry(entryId uuid.UUID, status string, completed bool) error {
    log.Debug(fmt.Sprintf("updating mail entry %+v with status '%s'", entryId, status))

    query := "UPDATE mail_relay_entries SET status=$1, completed=$2 WHERE entry_id=$3"
    _, err := db.Session.Exec(context.Background(), query, status, completed, entryId)
    if err != nil {
        log.Error(fmt.Errorf("unable to update mail entry: %+v", err))
    }
    return err
}

func(db *Persistence) GetMailEntries() ([]MailRelayEntry, error) {
    log.Debug("retrieving mail entries")
    entries := []MailRelayEntry{}

    query := `SELECT entry_id,event_timestamp,status,completed,data
    FROM mail_relay_entries`

    rows, err := db.Session.Query(context.Background(), query)
    if err != nil {
        switch err {
        case pgx.ErrNoRows:
            return entries, nil
        default:
            return entries, err
        }
    }

    for rows.Next() {
        var (entryId uuid.UUID; eventTimestamp time.Time)
        var (status string; completed bool; data map[string]interface{})

        if err := rows.Scan(&entryId, &eventTimestamp, &status,
            &completed, &data); err != nil {
            log.Warn(fmt.Errorf("unable to scan data into local variables: %+v", err))
            continue
        }
        entries = append(entries, MailRelayEntry{
            EntryId: entryId,
            EventTimestamp: eventTimestamp,
            Status: status,
            Completed: completed,
            Data: data,
        })
    }
    return entries, nil
}