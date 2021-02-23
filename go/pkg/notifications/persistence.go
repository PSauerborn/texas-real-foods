package notifications

import (
    "fmt"
    "context"
    "encoding/json"

    "github.com/jackc/pgx/v4"
    "github.com/google/uuid"
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

// function used to generate a new notification
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

type Notification struct {
    NotificationId uuid.UUID        `json:"notification_id"`
    Notification ChangeNotification `json:"notification"`
}

// function used to retrieve list of notifications
func(db *Persistence) GetNotifications() ([]Notification, error) {
    log.Debug("retrieving notifications from database...")

    notifications := []Notification{}
    query := `SELECT notification,notification_id FROM notifications`
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
        var (payloadJson []byte; notificationId uuid.UUID)
        if err := rows.Scan(&payloadJson, &notificationId); err != nil {
            log.Error(fmt.Errorf("unable to retreive notification from database: %+v", err))
            continue
        }
        // convert JSON format of notification into notification
        var notification ChangeNotification
        if err := json.Unmarshal(payloadJson, &notification); err != nil {
            log.Error(fmt.Errorf("unable to convert JSON to struct: %+v", err))
            continue
        }
        notifications = append(notifications, Notification{notificationId, notification})
    }
    return notifications, nil
}

// function used to retrieve list of notifications that are marked as unread
func(db *Persistence) GetUnreadNotifications() ([]Notification, error) {
    log.Debug("retrieving unread notifications from database...")

    notifications := []Notification{}
    query := `SELECT notification,notification_id FROM notifications
        WHERE read=false`
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
        var (payloadJson []byte; notificationId uuid.UUID)
        if err := rows.Scan(&payloadJson, &notificationId); err != nil {
            log.Error(fmt.Errorf("unable to retreive notification from database: %+v", err))
            continue
        }
        // convert JSON format of notification into notification
        var notification ChangeNotification
        if err := json.Unmarshal(payloadJson, &notification); err != nil {
            log.Error(fmt.Errorf("unable to convert JSON to struct: %+v", err))
            continue
        }
        notifications = append(notifications, Notification{notificationId, notification})
    }
    return notifications, nil
}

// function used to check if notification already exists. note that
// all notifications are stored with a hash to identify a notification
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

// function used to check if notification exists
func(db *Persistence) NotificationExists(notificationId uuid.UUID) (bool, error) {
    log.Debug(fmt.Sprintf("checking if notification %s exists...", notificationId))
    query := `SELECT hash FROM notifications WHERE notification_id=$1`
    row := db.Session.QueryRow(context.Background(), query, notificationId)
    var notificationHash string
    if err := row.Scan(&notificationHash); err != nil {
        switch err {
        case pgx.ErrNoRows:
            return false, nil
        default:
            return true, err
        }
    }
    return true, nil
}

// function used to set notification to read
func(db *Persistence) UpdateNotification(notificationId uuid.UUID) error {
    log.Debug(fmt.Sprintf("updating notification %s...", notificationId))
    query := `UPDATE notifications SET read=true WHERE notification_id=$1`
    _, err := db.Session.Exec(context.Background(), query, notificationId)
    if err != nil {
        log.Error(fmt.Errorf("unable to update notification: %+v", err))
        return err
    }
    return nil
}