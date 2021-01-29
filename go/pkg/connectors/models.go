package connectors

import (
    "github.com/google/uuid"
)

// struct to store business data. each update contains
// the business metadata to identify the business, and
// the updated business data (with source to identify the
// connector used to get the data)
type BusinessUpdate struct{
    Meta BusinessMetadata
    Data BusinessData
}

// struct to store business metadata
type BusinessMetadata struct{
    BusinessName   string                 `json:"business_name"`
    BusinessId	   uuid.UUID              `json:"business_id"`
    BusinessURI    string                 `json:"business_uri"`
    Metadata       map[string]interface{} `json:"metadata"`
}

// struct used to store updated business data values
type BusinessData struct{
    WebsiteLive    bool     `json:"website_live"`
    BusinessPhones []string `json:"business_phones"`
    Source         string   `json:"string"`
}