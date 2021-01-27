package api

import (
    "time"

    "github.com/google/uuid"
)

type NewAssetRequest struct{
    BusinessName string                 `json:"business_name"  validate:"required"`
    BusinessURI  string                 `json:"business_uri"   validate:"required"`
    Metadata     map[string]interface{} `json:"metadata"`
}

type AssetUpdateRequest struct{
    BusinessURI  string            `json:"business_uri"  validate:"required"`
    BusinessMeta map[string]string `json:"business_meta" validate:"required"`
}

type BusinessInfo struct{
    BusinessName   string                 `json:"business_name"`
    BusinessId	   uuid.UUID              `json:"business_id"`
    BusinessURI    string                 `json:"business_uri"`
    WebsiteLive    bool                   `json:"website_live"`
    BusinessPhones []string               `json:"business_phones"`
    LastUpdate     time.Time              `json:"last_update"`
    Added 		   time.Time              `json:"added"`
    Metadata       map[string]interface{} `json:"metadata"`
}