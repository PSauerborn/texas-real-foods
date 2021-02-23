package utils

import (
    "fmt"
    "time"
    "encoding/json"

    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
    "texas_real_foods/pkg/connectors"
)

type TexasRealFoodsAPIAccessor struct {
    *utils.BaseAPIAccessor
}

// function to generate new API accessor for Texas Real Foods API
func NewTRFApiAccessor(host, protocol string, port *int) *TexasRealFoodsAPIAccessor {
    baseAccessor := utils.BaseAPIAccessor{
        Host: host,
        Port: port,
        Protocol: protocol,
    }
    return &TexasRealFoodsAPIAccessor{
        &baseAccessor,
    }
}

// function to generate new API accessor for Texas Real Foods API
func NewTRFApiAccessorFromConfig(config utils.APIDependencyConfig) *TexasRealFoodsAPIAccessor {
    baseAccessor := utils.NewAPIAccessorFromConfig(config)
    return &TexasRealFoodsAPIAccessor{
        baseAccessor,
    }
}

type ListBusinessResponse struct {
    HTTPCode int 			               `json:"http_code"`
    Data     []connectors.BusinessMetadata `json:"data"`
}

// function to get businesses from texas real foods API
func(accessor *TexasRealFoodsAPIAccessor) GetBusinesses() (ListBusinessResponse, error) {
    log.Debug("fetching businesses from texas real foods api...")
    url := accessor.FormatURL("texas-real-foods/businesses")

    var response ListBusinessResponse
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved businesses from API"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    default:
        log.Error(fmt.Sprintf("failed to retrieve businesses: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}

type TimeseriesDataEntry struct {
    EventTimestamp time.Time
    connectors.BusinessData
}

type TimeseriesResponse struct {
    HTTPCode int 							  `json:"http_code"`
    Data     map[string][]TimeseriesDataEntry `json:"data"`
}

// function to get timeseries data for a given business from
// texas real foods API with a given business ID, start and
// end time
func(accessor *TexasRealFoodsAPIAccessor) GetTimeseriesData(businessId uuid.UUID,
    start, end time.Time) (TimeseriesResponse, error) {
    log.Debug(fmt.Sprintf("fetching timeseries data for business %s in range %s - %s...",
        businessId, start, end))

    // format timestamps to match required datetime format
    startTs := start.Format("2006-01-02T15:04")
    endTs := end.Format("2006-01-02T15:04")
    // construct URL using parameters
    url := accessor.FormatURL(fmt.Sprintf("texas-real-foods/data/timeseries/%s/%s/%s",
        businessId, startTs, endTs))

    var response TimeseriesResponse
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved timeseries data from API"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    default:
        log.Error(fmt.Sprintf("failed to retrieve timeseries data: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}

// function to get timeseries data for a given business from
// texas real foods API with a given business ID, start and
// end time
func(accessor *TexasRealFoodsAPIAccessor) GetTimeseriesDataCounted(businessId uuid.UUID,
    count int) (TimeseriesResponse, error) {
    log.Debug(fmt.Sprintf("fetching timeseries data for business %s with limit %d...",
        businessId, count))

    // construct URL using parameters
    url := accessor.FormatURL(fmt.Sprintf("texas-real-foods/data/timeseries-count/%s/%d",
        businessId, count))

    var response TimeseriesResponse
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved timeseries data from API"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    default:
        log.Error(fmt.Sprintf("failed to retrieve timeseries data: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}

type StaticDataResponse struct {
    HTTPCode int 					   `json:"http_code"`
    Data     []connectors.BusinessData `json:"data"`
}

// function to get static data from texas real foods API for
// a given business with business ID
func(accessor *TexasRealFoodsAPIAccessor) GetStaticData(businessId uuid.UUID,
    start, end time.Time) (interface{}, error) {
    log.Debug(fmt.Sprintf("fetching static data for business %s...", businessId))

    url := accessor.FormatURL(fmt.Sprintf("texas-real-foods/data/static/%s",
        businessId))

    var response StaticDataResponse
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("GET", url, nil, nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response, err
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully retrieved static data from API"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response, err
        }
        return response, nil
    default:
        log.Error(fmt.Sprintf("failed to retrieve static data: received invalid %d response from API",
            resp.StatusCode))
        return response, utils.ErrInvalidAPIResponse
    }
}
