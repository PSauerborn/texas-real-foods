package notifications

import (
    "fmt"
    "time"
    "errors"

    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)

var (
    ErrHashAlreadyExists = errors.New("Notification hash already exists")
)


type ChangeNotification struct{
    BusinessId       uuid.UUID              `json:"business_id" binding:"required"`
    BusinessName     string                 `json:"business_name" binding:"required"`
    EventTimestamp   time.Time              `json:"event_timestamp" binding:"required"`
    Notification     string                 `json:"notification" binding:"required"`
    NotificationHash string                 `json:"notification_hash" binding:"required"`
    Metadata         map[string]interface{} `json:"metadata"`
}

// define interface for engine
type NotificationEngine interface {
    SendNotification(notification ChangeNotification) error
}

func NewDefaultNotificationEngine(postgresUrl string) *DefaultNotificationEngine {
    return &DefaultNotificationEngine{postgresUrl}
}

type DefaultNotificationEngine struct {
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
    defer conn.Close()

    // check if notification hash already exists to prevent duplicate notifications
    exists, err := db.NotificationHashExists(notification.NotificationHash)
    if err != nil {
        log.Error(fmt.Errorf("unable to verify notification hash: %+v", err))
        return err
    }

    // return error if hash of notification already exists
    if exists {
        log.Warn("notification hash already exists. skipping sending of notification...")
        return ErrHashAlreadyExists
    }
    return db.CreateNotification(notification)
}