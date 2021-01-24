package utils

import (
    "fmt"
    "strings"
    "errors"
    "regexp"
    "bytes"
    "net/http"
    "encoding/json"

    log "github.com/sirupsen/logrus"
)

var (
    // define mappings for phone number regexes
    PhoneRegexMapping = map[string]*regexp.Regexp{
        "uk-1": regexp.MustCompile(`((\(?0\d{4}\)?\s?\d{3}\s?\d{3})|(\(?0\d{3}\)?\s?\d{3}\s?\d{4})|(\(?0\d{2}\)?\s?\d{4}\s?\d{4}))(\s?\#(\d{4}|\d{3}))?`),
        "uk-2": regexp.MustCompile(`(\+44\s?7\d{3}|\(?07\d{3}\)?)\s?\d{3}\s?\d{3}`),
        "uk-3": regexp.MustCompile(`(((\+44\s?\d{4}|\(?0\d{4}\)?)\s?\d{3}\s?\d{3})|((\+44\s?\d{3}|\(?0\d{3}\)?)\s?\d{3}\s?\d{4})|((\+44\s?\d{2}|\(?0\d{2}\)?)\s?\d{4}\s?\d{4}))(\s?\#(\d{4}|\d{3}))?`),
        "us-1": regexp.MustCompile(`[2-9]\d{2}-\d{3}-\d{4}`),
        "us-2": regexp.MustCompile(`((\(\d{3}\)?)|(\d{3}))([\s-./]?)(\d{3})([\s-./]?)(\d{4})`),
        "us-3": regexp.MustCompile(`\(?[\d]{3}\)?[\s-]?[\d]{3}[\s-]?[\d]{4}`),
    }

    // define custom errors
    ErrInvalidAPIResponse = errors.New("Received invalid response from phone API")
)

// helper function used to check if a string slice contains
// a particular string value
func StringSliceContains(str string, items []string) bool {
    for _, item := range(items) {
        if str == item {
            return true
        }
    }
    return false
}

// helper function used to clean common punctuations from
// a phone number to prevent duplicate numbers
func CleanNumber(number string) string {
    var cleaned string
    // remove all common string occurrences
    cleaned = strings.ReplaceAll(number, "-", "")
    cleaned = strings.ReplaceAll(cleaned, " ", "")
    cleaned = strings.ReplaceAll(cleaned, "(", "")
    cleaned = strings.ReplaceAll(cleaned, ")", "")
    cleaned = strings.ReplaceAll(cleaned, "+", "")
    return cleaned
}

// helper function used to parse contents of a string
// and search for phone numbers by regex matches
func GetPhoneNumbersByRegex(text string) []string {
    matches := []string{}
    // iterate over regexes and find matches
    for code, exp := range(PhoneRegexMapping) {
        log.Debug(fmt.Sprintf("checking regex match for code '%s'", code))
        match := exp.FindAllString(text, -1)

        // iterate over matches and append to results
        for _, matchString := range(match) {
            // remove all punctuation from phone numbers to remove duplicates
            cleanedMatch := CleanNumber(matchString)
            if len(matchString) > 0 && !(StringSliceContains(cleanedMatch, matches)) {
                log.Debug(fmt.Sprintf("found phone number match %s", cleanedMatch))
                matches = append(matches, cleanedMatch)
            }
        }
    }
    log.Info(fmt.Sprintf("found phone number matches for numbers %+v", matches))
    return matches
}


type PhoneNumberValidationResponse struct{
    HTTPCode int                          `json:"http_code"`
    Data     PhoneNumberValidationResults `json:"data"`
}

type PhoneNumberValidationResults struct{
    Valid   []string `json:"valid"`
    Invalid []string `json:"invalid"`
}

type PhoneNumberValidationRequest struct{
    CountryCode string   `json:"country_code"`
    Numbers     []string `json:"numbers"`
}

// function used to validate phone numbers against phonenumber validation API
func ValidatePhoneNumbers(apiHost string, numbers []string) (PhoneNumberValidationResults, error) {
    log.Debug(fmt.Sprintf("validating phone numbers against API"))

    request := PhoneNumberValidationRequest{
        CountryCode: "US",
        Numbers: numbers,
    }
    // convert payload to JSON format
    jsonData, err := json.Marshal(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to convert numbers to JSON: %+v", err))
        return PhoneNumberValidationResults{}, err
    }
    req, err := http.NewRequest("POST", fmt.Sprintf("%s/validate", apiHost), bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")

    // generate new HTTP client and execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request"))
        return PhoneNumberValidationResults{}, err
    }
    defer resp.Body.Close()

    // if API request was successful, parse JSON body and response
    if resp.StatusCode != 200 {
        log.Error(fmt.Sprintf("received invalid API response with code %d", resp.StatusCode))
        return PhoneNumberValidationResults{}, ErrInvalidAPIResponse
    }

    // convert response to JSON format
    var response PhoneNumberValidationResponse
    if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
        log.Error(fmt.Errorf("unable to parse JSON response from API"))
        return PhoneNumberValidationResults{}, ErrInvalidAPIResponse
    }
    return response.Data, nil
}