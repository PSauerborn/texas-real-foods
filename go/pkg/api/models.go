package api

import (
    "time"

    "github.com/google/uuid"
)

type NewBusinessRequest struct{
    BusinessName string                 `json:"business_name" binding:"required"`
    BusinessURI  string                 `json:"business_uri"  binding:"required"`
    Metadata     map[string]interface{} `json:"metadata"`
}

type BusinessUpdateRequest struct{
    BusinessURI string `json:"business_uri" binding:"required"`
}

type BusinessMetaPatchRequest struct{
    Operation []map[string]interface{} `json:"operation" binding:"required"`
}

type BusinessInfo struct{
    BusinessName   string                 `json:"business_name"`
    BusinessId	   uuid.UUID              `json:"business_id"`
    BusinessURI    string                 `json:"business_uri"`
    LastUpdate     time.Time              `json:"last_update"`
    Added 		   time.Time              `json:"added"`
    Metadata       map[string]interface{} `json:"metadata"`
}

type Notification struct{
    NotificationId uuid.UUID              `json:"notification_id"`
    EventTimestamp time.Time              `json:"event_timestamp"`
    Notification   map[string]interface{} `json:"notification"`
    Hash           string                 `json:"hash"`
}