package utils

import (
    "fmt"
    "bytes"
    "encoding/json"

    log "github.com/sirupsen/logrus"

    "texas_real_foods/pkg/utils"
)


type UtilsAPIAccessor struct {
    *utils.BaseAPIAccessor
}

func NewUtilsAPIAccessor(host, protocol string, port *int) *UtilsAPIAccessor {
    baseAccessor := utils.BaseAPIAccessor{
        Host: host,
        Port: port,
        Protocol: protocol,
    }
    return &UtilsAPIAccessor{
        &baseAccessor,
    }
}

// function to generate new API accessor for Texas Real Foods API
func NewUtilsAPIAccessorFromConfig(config utils.APIDependencyConfig) *UtilsAPIAccessor {
    baseAccessor := utils.NewAPIAccessorFromConfig(config)
    return &UtilsAPIAccessor{
        baseAccessor,
    }
}

type PhoneNumberValidationResponse struct {
    HTTPCode int                          `json:"http_code"`
    Data     PhoneNumberValidationResults `json:"data"`
}

type PhoneNumberValidationResults struct {
    Valid   []string `json:"valid"`
    Invalid []string `json:"invalid"`
}

type PhoneNumberValidationRequest struct {
    CountryCode string   `json:"country_code"`
    Numbers     []string `json:"numbers"`
}

// function used to validate phone numbers against phone
// validation API
func(accessor *UtilsAPIAccessor) ValidatePhoneNumbers(numbers []string) (
    PhoneNumberValidationResults, error) {
    log.Debug(fmt.Sprintf("validating %d numbers against utils API", len(numbers)))
    url := accessor.FormatURL("validate")

    var response PhoneNumberValidationResponse
    // serialize request body into JSON string
    jsonBody, err := json.Marshal(PhoneNumberValidationRequest{
        CountryCode: "US",
        Numbers: numbers})
    if err != nil {
        return response.Data, utils.ErrInvalidRequestBodyJSON
    }
    // generate new JSON request and execute
    req, err := accessor.NewJSONRequest("POST", url, bytes.NewBuffer(jsonBody), nil)
    if err != nil {
        log.Error(fmt.Errorf("unable to create request: %+v", err))
        return response.Data, err
    }
    resp, err := accessor.ExecuteRequest(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute API request: %+v", err))
        return response.Data, err
    }
    defer resp.Body.Close()

    // handle response based on code
    switch resp.StatusCode {
    case 200:
        log.Debug(fmt.Sprintf("successfully validated phone numbers from API"))
        // decode JSON response from API and return
        if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
            log.Error(fmt.Errorf("unable to parse JSON response from API: %+v", err))
            return response.Data, err
        }
        return response.Data, nil
    default:
        log.Error(fmt.Sprintf("failed to validate phone numbers: received invalid %d response from API",
            resp.StatusCode))
        return response.Data, utils.ErrInvalidAPIResponse
    }
}