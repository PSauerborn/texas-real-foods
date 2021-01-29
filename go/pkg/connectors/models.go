package connectors

import (
    "github.com/google/uuid"
)


type BusinessUpdate struct{
    Meta BusinessMetadata
    Data BusinessData
}

type BusinessMetadata struct{
    BusinessName   string                 `json:"business_name"`
    BusinessId	   uuid.UUID              `json:"business_id"`
    BusinessURI    string                 `json:"business_uri"`
    Metadata       map[string]interface{} `json:"metadata"`
}

type BusinessData struct{
    WebsiteLive    bool     `json:"website_live"`
    BusinessPhones []string `json:"business_phones"`
    Source         string   `json:"string"`
}