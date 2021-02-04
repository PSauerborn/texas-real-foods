package mail_relay

import (
	"time"

	"github.com/google/uuid"
)

type ZipCodeData struct {
	EconomicRegion string   `json:"economic_region"`
	County		   string   `json:"county"`
	State		   string   `json:"state"`
	ZipCode        string   `json:"zip_code"`
	AreaCodes	   []string `json:"area_codes"`
}

type ZipCodeDataResponse struct {
	HTTPCode int         `json:"http_code"`
	Data     ZipCodeData `json:"data"`
}

type MailRelayRequest struct {
    FirstName string    `json:"first_name" binding:"required"`
    LastName  string    `json:"last_name"  binding:"required"`
	ZipCode   string    `json:"zip_code"   binding:"required"`
	EntryId   uuid.UUID `json:"entry_id"`
}

type MailRelayEntry struct {
	EntryId        uuid.UUID 			  `json:"entry_id"`
	EventTimestamp time.Time 			  `json:"event_timestamp"`
	Status   	   string  			      `json:"status"`
	Completed 	   bool  				  `json:"completed"`
	Data 		   map[string]interface{} `json:"data"`
}