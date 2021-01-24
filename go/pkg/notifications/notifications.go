package notifications

import (
    "fmt"
    "time"
    "context"

    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)


type ChangeNotification struct{
    BusinessId     uuid.UUID `json:"business_id"`
    BusinessName   string    `json:"business_name"`
    EventTimestamp time.Time `json:"event_timestamp"`
    Notification   string    `json:"notification"`
}

// define interface for engine
type NotificationEngine interface{
    SendNotification(notification ChangeNotification) error
}

func NewDefaultNotificationEngine(postgresUrl string) *DefaultNotificationEngine {
    return &DefaultNotificationEngine{postgresUrl}
}

type DefaultNotificationEngine struct{
    PostgresURL string
}

func(e *DefaultNotificationEngine) SendNotification(notification ChangeNotification) error {
    // establish new connection to postgres persistence
    db := NewPersistence(e.PostgresURL)
    conn, err := db.Connect()
    if err != nil {
        log.Error(fmt.Errorf("unable to connect notification engine to postgres: %+v", err))
        return err
    }
    defer conn.Close(context.Background())
    return db.CreateNotification(notification)
}