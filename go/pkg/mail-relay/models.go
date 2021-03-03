package mail_relay

import (
    "time"

    "github.com/google/uuid"
)

type MailRelayRequest struct {
    FirstName   string    `json:"first_name"    binding:"required"`
    LastName    string    `json:"last_name"     binding:"required"`
    ZipCode     string    `json:"zip_code"      binding:"required"`
    Email       string    `json:"email"         binding:"required"`
    DateOfBirth string    `json:"date_of_birth" binding:"required"`
    EntryId     uuid.UUID `json:"entry_id"`
}

type MailRelayEntry struct {
    EntryId        uuid.UUID 			  `json:"entry_id"`
    EventTimestamp time.Time 			  `json:"event_timestamp"`
    Status   	   string  			      `json:"status"`
    Completed 	   bool  				  `json:"completed"`
    Data 		   map[string]interface{} `json:"data"`
}