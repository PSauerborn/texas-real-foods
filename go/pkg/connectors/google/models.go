package connectors

import (
    "fmt"
    "errors"
    "encoding/json"

    "github.com/go-playground/validator/v10"
    log "github.com/sirupsen/logrus"
)

var (
    // define custom errors
    ErrInvalidGoogleMetadata = errors.New("Invalid google metadata")

    // create new validator
    validate = validator.New()
)


// function to parse google metadata from business metadata, note
// that businesses might not have a google metadata entry in the
// postgres database, in which case an error is returned
func ParseGoogleMetadata(data map[string]interface{}) (GoogleMetadata, error) {
    log.Debug(fmt.Sprintf("converting %+v to google metadata", data))

    // convert to JSON string to parse struct
    jsonString, err := json.Marshal(data)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert map to google metadata"))
        return GoogleMetadata{}, ErrInvalidGoogleMetadata
    }

    var meta GoogleMetadata
    // unmarshal JSON string into struct; return error if not possible
    if err := json.Unmarshal(jsonString, &meta); err != nil {
        log.Error(fmt.Errorf("unable to cast metadata to google format: %+v", err))
        return GoogleMetadata{}, ErrInvalidGoogleMetadata
    }
    // return struct (and validate to catch any errors)
    return meta, validate.Struct(meta)
}

type GoogleMetadata struct {
    GooglePlaceId string `json:"google_place_id" validate:"required"`
}

type GoogleAPIResponse struct {
    Name                 string `json:"name"`
    FormattedAddress     string `json:"formatted_address"`
    FormattedPhoneNumber string `json:"formatted_phone_number"`
    PermanentlyClosed    string `json:"permanently_closed"`
    PlaceId              string `json:"place_id"`
    Website              string `json:"website"`
}