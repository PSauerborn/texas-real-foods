package connectors

import (
	"time"

	"github.com/google/uuid"
)


// define struct used to store basic business information
type BusinessInfo struct{
	BusinessName   string    `json:"business_name"`
	BusinessId	   uuid.UUID `json:"business_id"`
	BusinessURI    string    `json:"business_uri"`
	WebsiteLive    bool      `json:"website_live"`
	BusinessPhones []string  `json:"business_phones"`
	LastUpdate     time.Time `json:"last_update"`
}
