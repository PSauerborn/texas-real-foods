package connectors

import (
    "fmt"
    "net/http"
    "io"
    "errors"
    "encoding/json"

    log "github.com/sirupsen/logrus"
)

var (
    // define base URL for yelp API
    baseApiURL = "https://api.yelp.com/v3/businesses"

    // define custom errors
    ErrInvalidAPIResponse  = errors.New("Received invalid API response")
    ErrBusinessNotFound    = errors.New("Cannot find API business entry")
    ErrUnauthorized        = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse = errors.New("Received invalid JSON response from yelp API")
    ErrRequestLimitReached = errors.New("Reached request limit on API")
)

// function used to parse yelp response
func ParseYelpResponse(content io.ReadCloser) (YelpBusinessResults, error) {
    var response YelpBusinessResponse
    if err := json.NewDecoder(content).Decode(&response); err != nil {
        log.Error(fmt.Sprintf("unable to parse API response into struct: %+v", err))
        return YelpBusinessResults{}, ErrInvalidJSONResponse
    }
    // convert data into struct
    data := YelpBusinessResults{
        BusinessId: response.Id,
        BusinessName: response.Name,
        PhoneNumber: response.Phone,
    }
    return data, nil
}

func GetYelpBusinessInfo(businessId, apiKey string) (YelpBusinessResults, error) {
    log.Debug(fmt.Sprintf("requesting Yelp! data for business %s", businessId))
    // createnew HTTP instance and set request headers
    req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s", baseApiURL, businessId), nil)
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return YelpBusinessResults{}, err
    }
    defer resp.Body.Close()

    // handle response based on status code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved asset data for asset '%s'", businessId))
        // parse response body and convert into struct
        results, err := ParseYelpResponse(resp.Body)
        if err != nil {
            log.Error(fmt.Sprintf("unable to parse JSON response: %+v", err))
            return YelpBusinessResults{}, ErrInvalidAPIResponse
        }
        return results, nil
    case 401:
        log.Error(fmt.Sprintf("received unauthorized response from yelp API"))
        return YelpBusinessResults{}, ErrUnauthorized
    case 404:
        log.Error(fmt.Sprintf("cannot find API results for business ID %s", businessId))
        return YelpBusinessResults{}, ErrBusinessNotFound
    case 429:
        log.Error("reached request limit on API")
        return YelpBusinessResults{}, ErrRequestLimitReached
    default:
        log.Error(fmt.Errorf("received invalid response from yelp API with code %d", resp.StatusCode))
        return YelpBusinessResults{}, ErrInvalidAPIResponse
    }
}