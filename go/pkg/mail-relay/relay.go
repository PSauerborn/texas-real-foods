package mail_relay

import (
    "fmt"
    "errors"
    "strings"
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"

    log "github.com/sirupsen/logrus"

    apis "texas_real_foods/pkg/utils/api_accessors"
)

var (
    // generate new event channel to process requests
    eventChannel = make(chan MailRelayRequest)

    // define custom errors
    ErrInvalidZipCodeResponse = errors.New("Received invalid response from zipcode API")
    ErrInvalidAPIResponse     = errors.New("Received invalid API response")
    ErrZipCodeNotFound        = errors.New("Cannot find zipcode entry in API")
    ErrUnauthorized           = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse    = errors.New("Received invalid JSON response from zipcode API")
    ErrRequestLimitReached    = errors.New("Reached request limit on API")
)

func TriggerMailChimpAsync(request MailRelayRequest, data apis.ZipCodeData) {
    if err := TriggerMailChimp(request, data); err != nil {
        // insert failed message into event log
        persistence.UpdateMailEntry(request.EntryId, "failed", false)
    } else {
        // insert success message into event log
        persistence.UpdateMailEntry(request.EntryId, "completed", true)
    }
}

// function used to add a new member to a given mail chimp list
func TriggerMailChimp(request MailRelayRequest, data apis.ZipCodeData) error {
    // function used to relay sign up request to mail chimp server
    log.Info(fmt.Sprintf("relaying request %+v", request))
    mailChimpRequest := map[string]interface{}{
        "email_address": request.Email,
        "status": "subscribed",
        "email_type": "html",
        "merge_fields": map[string]string{
            "FNAME": request.FirstName,
            "LNAME": request.LastName,
            "ZIP": data.ZipCode,
            "CB": "",
            "DATEB": request.DateOfBirth,
            "COUNTY": strings.Replace(data.County, " County", "", -1),
            "METRO": "",
            "ECON": data.EconomicRegion,
        },
        "vip": false,
        "language": "en",
        "tags": []string{},
        "source": "API - Mail Relay",
    }
    log.Debug(fmt.Sprintf("making request to mail chimp server with body %+v", mailChimpRequest))
    // convert request to JSON format and add to request
    jsonRequest, err := json.Marshal(mailChimpRequest)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert mail chimp request to JSON format: %+v", err))
        return err
    }
    // generate url based on list ID and base API url
    url := fmt.Sprintf("%s/list/%s/members", mailChimpConfig.APIUrl, mailChimpConfig.ListID)
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonRequest))
    if err != nil {
        log.Error(fmt.Errorf("unable to trigger mail chimp request: %+v", err))
        return err
    }
    // set basic auth on request
    req.SetBasicAuth("psauerborn", mailChimpConfig.APIKey)

    client := http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to trigger mail chimp request: %+v", err))
        return err
    }
    defer resp.Body.Close()

    switch resp.StatusCode {
    case 200:
        log.Info("successfully triggered mail chimp request")
        return nil
    default:
        body, _ := ioutil.ReadAll(resp.Body)
        log.Error(fmt.Errorf("unable to trigger mail chimp request: received response %s", string(body)))
        return ErrInvalidAPIResponse
    }
}

// function used to get zip code data from utils API
func GetZipCodeData(zipcode string) (apis.ZipCodeData, error) {
    log.Info(fmt.Sprintf("requesting zip code data for code '%s'", zipcode))
    accessor := apis.NewUtilsAPIAccessorFromConfig(utilsAPIConfig)
    var data apis.ZipCodeData
    data, err := accessor.GetZipCodeData(zipcode)
    if err != nil {
        log.Error(fmt.Errorf("unable to retrieve zip code data for code %s: %+v", zipcode, err))
        return data, err
    }
    log.Debug(fmt.Sprintf("successfully retrieve zip code data %+v", data))
    return data, nil
}
