package auto_updater

import (
    "io"
    "fmt"
    "errors"
    "encoding/json"
    "net/http"

    log "github.com/sirupsen/logrus"
)

var (
    // define custom errors
    ErrInvalidAPIResponse  = errors.New("Received invalid API response")
    ErrBusinessNotFound    = errors.New("Cannot find API business entry")
    ErrUnauthorized        = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse = errors.New("Received invalid JSON response from API")
    ErrRequestLimitReached = errors.New("Reached request limit on API")
)

// function used to parse response from data API
func ParseAPIResponse(data io.ReadCloser) (ListBusinessResponse, error) {
    var response ListBusinessResponse
    if err := json.NewDecoder(data).Decode(&response); err != nil {
        return response, ErrInvalidJSONResponse
    }
    log.Debug(fmt.Sprintf("successfully extracted data response %+v", response))
    return response, nil
}

func GetBusinessesFromAPI(baseApiUrl string) (ListBusinessResponse, error) {
    log.Debug("retrieving businesses from Texas Real Foods API")
    // createnew HTTP instance and set request headers
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/businesses", baseApiUrl), nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP Request: %+v", err))
        return ListBusinessResponse{}, err
    }
    // set JSON as content type
    req.Header.Set("Content-Type", "application/json")

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return ListBusinessResponse{}, err
    }
    defer resp.Body.Close()

    // handle response based on status code
    switch resp.StatusCode {
    case 200:
        log.Debug("successfully retrieved business data")
        // parse response body and convert into struct
        results, err := ParseAPIResponse(resp.Body)
        if err != nil {
            log.Error(fmt.Sprintf("unable to parse JSON response: %+v", err))
            return results, ErrInvalidAPIResponse
        }
        return results, nil
    default:
        log.Error(fmt.Errorf("received invalid response from data API with code %d", resp.StatusCode))
        return ListBusinessResponse{}, ErrInvalidAPIResponse
    }
}