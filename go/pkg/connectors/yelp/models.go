package connectors

import (
	"fmt"
	"errors"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

var (
	ErrInvalidYelpMetadata = errors.New("Invalid yelp metadata")

	// create new validator
	validate = validator.New()
)


func ParseYelpMetadata(data map[string]interface{}) (YelpMetadata, error) {
	log.Debug(fmt.Sprintf("converting %+v to yelp metadata", data))

	jsonString, err := json.Marshal(data)
	if err != nil {
		log.Error(fmt.Errorf("unable to convert map to yelp metadata"))
		return YelpMetadata{}, ErrInvalidYelpMetadata
	}

	var meta YelpMetadata
	if err := json.Unmarshal(jsonString, &meta); err != nil {
		log.Error(fmt.Errorf("unable to cast metadata to yelp format: %+v", err))
		return YelpMetadata{}, ErrInvalidYelpMetadata
	}
	return meta, validate.Struct(meta)
}

type YelpMetadata struct{
	YelpBusinessId string `json:"yelp_business_id" validate:"required"`
}

type GeoCoordinates struct{
	Latitude  float64 `json:"latitude"  validate:"required"`
	Longitude float64 `json:"longitude" validate:"required"`
}

type YelpBusinessResults struct{
	BusinessId   string `json:"business_id"`
	BusinessName string `json:"business_name"`
	PhoneNumber  string `json:"phone_number"`
}

type YelpBusinessResponse struct{
	Id          string         `json:"id"`
	Name        string         `json:"name"`
	Phone       string         `json:"phone"`
	IsClosed    bool           `json:"is_closed"`
	Coordinates GeoCoordinates `json:"coordinates"`
}
