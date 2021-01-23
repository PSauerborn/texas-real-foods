package api

import (
	"time"

	"github.com/google/uuid"
)

type NewAssetRequest struct{
	BusinessName  string    `json:"business_name"  validate:"required"`
	BusinessURI   string    `json:"business_uri"   validate:"required"`
}

type BusinessInfo struct{
	BusinessName   string    `json:"business_name"`
	BusinessId	   uuid.UUID `json:"business_id"`
	BusinessURI    string    `json:"business_uri"`
	WebsiteLive    bool      `json:"website_live"`
	BusinessPhones []string  `json:"business_phones"`
	LastUpdate     time.Time `json:"last_update"`
	Added 		   time.Time `json:"added"`
}