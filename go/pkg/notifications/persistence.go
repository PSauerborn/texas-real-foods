package notifications

import (
    "fmt"
    "context"
    "time"
    "encoding/json"

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

func(db *Persistence) CreateNotification(payload ChangeNotification) error {
    log.Debug(fmt.Sprintf("storing notification %+v", payload))

    // convert payload to JSON format and store
    payloadJson, _ := json.Marshal(payload)
    notificationId := uuid.New()

    query := `INSERT INTO notifications(notification_id,notification,hash) VALUES($1,$2,$3)`
    _, err := db.Session.Exec(context.Background(), query, notificationId, payloadJson,
        payload.NotificationHash)
    return err
}

func(db *Persistence) GetNotifications() ([]ChangeNotification, error) {
    log.Debug("retrieving notifications from database")

    notifications := []ChangeNotification{}
    query := `SELECT notification_id,event_timestamp,notification FROM notifications`
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
        var (notificationId uuid.UUID; eventTimestamp time.Time; payloadJson []byte;)
        if err := rows.Scan(&notificationId, &eventTimestamp, &payloadJson); err != nil {
            log.Error(fmt.Errorf("unable to retreive notification from database: %+v", err))
            continue
        }

        // convert JSON format of notification into notification
        var notification ChangeNotification
        if err := json.Unmarshal(payloadJson, &notification); err != nil {
            log.Error(fmt.Errorf("unable to convert JSON to struct: %+v", err))
            continue
        }

        notifications = append(notifications, notification)
    }
    return notifications, nil
}

func(db *Persistence) NotificationHashExists(hashed string) (bool, error) {
    log.Debug(fmt.Sprintf("checking notifications for hash %s", hashed))

    query := `SELECT notification_id FROM notifications WHERE hash=$1`
    row := db.Session.QueryRow(context.Background(), query, hashed)

    var notificationId uuid.UUID
    if err := row.Scan(&notificationId); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return false, nil
        default:
            return true, err
        }
    }
    return true, nil
}