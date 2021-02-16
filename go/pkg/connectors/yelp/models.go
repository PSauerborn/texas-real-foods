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
    ErrInvalidYelpMetadata = errors.New("Invalid yelp metadata")

    // create new validator
    validate = validator.New()
)

// function to parse yelp metadata from business metadata, note
// that businesses might not have a yelp metadata entry in the
// postgres database, in which case an error is returned
func ParseYelpMetadata(data map[string]interface{}) (YelpMetadata, error) {
    log.Debug(fmt.Sprintf("converting %+v to yelp metadata", data))

    // convert to JSON string to parse struct
    jsonString, err := json.Marshal(data)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert map to yelp metadata"))
        return YelpMetadata{}, ErrInvalidYelpMetadata
    }

    var meta YelpMetadata
    // unmarshal JSON string into struct; return error if not possible
    if err := json.Unmarshal(jsonString, &meta); err != nil {
        log.Error(fmt.Errorf("unable to cast metadata to yelp format: %+v", err))
        return YelpMetadata{}, ErrInvalidYelpMetadata
    }
    // return struct (and validate to catch any errors)
    return meta, validate.Struct(meta)
}

// struct to store metadata required to make Yelp requests
type YelpMetadata struct{
    YelpBusinessId string `json:"yelp_business_id" validate:"required"`
}

// struct used to store geo coordinates
type GeoCoordinates struct{
    Latitude  float64 `json:"latitude"  validate:"required"`
    Longitude float64 `json:"longitude" validate:"required"`
}

// struct to store values retrieved from yelp API
type YelpBusinessResults struct{
    BusinessId   string `json:"business_id"`
    BusinessName string `json:"business_name"`
    PhoneNumber  string `json:"phone_number"`
    IsOpen       bool   `json:"is_open"`
}

// struct to store API response body from yelp
type YelpBusinessResponse struct{
    Id          string         `json:"id"`
    Name        string         `json:"name"`
    Phone       string         `json:"phone"`
    IsClosed    bool           `json:"is_closed"`
    Coordinates GeoCoordinates `json:"coordinates"`
}
