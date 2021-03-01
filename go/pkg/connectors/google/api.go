package connectors

import (
    "io"
    "fmt"
    "errors"
    "encoding/json"
    "net/http"

    log "github.com/sirupsen/logrus"
)

var (
    // define base URL for google API
    baseApiURL = "https://maps.googleapis.com/maps/api/place/details/json"

    // define custom errors
    ErrInvalidAPIResponse  = errors.New("Received invalid API response")
    ErrBusinessNotFound    = errors.New("Cannot find API business entry")
    ErrUnauthorized        = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse = errors.New("Received invalid JSON response from google API")
    ErrRequestLimitReached = errors.New("Reached request limit on API")
)

// function to generate query string for Google Place API
func GenerateQueryString(apiKey, placeId string) string {
    fields := "formatted_address,name,permanently_closed,url,place_id,website,business_status,formatted_phone_number"
    return fmt.Sprintf("place_id=%s&fields=%s&key=%s", placeId, fields, apiKey)
}

// function used to parse response from google place API
func ParseGoogleResponse(data io.ReadCloser) (GoogleAPIResponse, error) {
    var response struct {
        Result GoogleAPIResponse `json:"result"`
        Status string            `json:"status"`
    }
    if err := json.NewDecoder(data).Decode(&response); err != nil {
        return response.Result, ErrInvalidJSONResponse
    }
    log.Debug(fmt.Sprintf("successfully extracted google response %+v", response))
    return response.Result, validate.Struct(response.Result)
}

// function used to get data
func GetGoogleBusinessInfo(placeId string, apiKey string) (GoogleAPIResponse, error) {
    log.Debug(fmt.Sprintf("making new request to Google API for ID '%s'", placeId))

    queryString := GenerateQueryString(apiKey, placeId)
    url := fmt.Sprintf("%s?%s", baseApiURL, queryString)
    // createnew HTTP instance and set request headers
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP Request: %+v", err))
        return GoogleAPIResponse{}, err
    }
    // set JSON as content type
    req.Header.Set("Content-Type", "application/json")

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return GoogleAPIResponse{}, err
    }
    defer resp.Body.Close()

    // handle response based on status code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved business data for asset '%s'", placeId))
        // parse response body and convert into struct
        results, err := ParseGoogleResponse(resp.Body)
        if err != nil {
            log.Error(fmt.Sprintf("unable to parse JSON response: %+v", err))
            return results, ErrInvalidAPIResponse
        }
        return results, nil
    case 401:
        log.Error(fmt.Sprintf("received unauthorized response from google API"))
        return GoogleAPIResponse{}, ErrUnauthorized
    case 404:
        log.Error(fmt.Sprintf("cannot find API results for business ID %s", placeId))
        return GoogleAPIResponse{}, ErrBusinessNotFound
    case 429:
        log.Error("reached request limit on API")
        return GoogleAPIResponse{}, ErrRequestLimitReached
    default:
        log.Error(fmt.Errorf("received invalid response from google API with code %d", resp.StatusCode))
        return GoogleAPIResponse{}, ErrInvalidAPIResponse
    }
}
