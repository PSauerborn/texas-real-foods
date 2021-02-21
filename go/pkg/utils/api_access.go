package utils

import (
    "io"
    "fmt"
    "errors"
    "time"
    "net/http"

    log "github.com/sirupsen/logrus"
)

var (
    // define custom errors
    ErrInvalidAPIResponse     = errors.New("Received invalid API response")
    ErrBusinessNotFound       = errors.New("Cannot find API business entry")
    ErrUnauthorized           = errors.New("Received unauthorized response from API")
    ErrInvalidJSONResponse    = errors.New("Received invalid JSON response from API")
    ErrRequestLimitReached    = errors.New("Reached request limit on API")
    ErrInvalidRequestBodyJSON = errors.New("Invalid request body: data must be JSON serializable")
)

type APIDependencyConfig struct {
    Name string
    Host string
    Port *int
    Protocol string
}

type BaseAPIAccessor struct {
    Host string
    Port *int
    Protocol string
}

// function used to execute a given request
func(accessor *BaseAPIAccessor) ExecuteRequest(request *http.Request) (*http.Response, error) {
    // generate new HTTP client and execute request
    start := time.Now()
    log.Debug(fmt.Sprintf("making request to url %s...", request.URL))
    client := &http.Client{}
    resp, err := client.Do(request)
    if err != nil {
        log.Error(fmt.Errorf("unable to execute HTTP request: %+v", err))
        return nil, err
    }
    // evaluate time elapsed to process request and log
    elapsed := time.Now().Sub(start)
    log.Info(fmt.Sprintf("processed request in %fs", elapsed.Seconds()))
    return resp, nil
}

// function to format url using a given protocol. host and port
// are inserted based on the values passed to the accessor
func(accessor *BaseAPIAccessor) FormatURL(url string) string {
    if accessor.Port != nil {
        return fmt.Sprintf("%s://%s:%d/%s", accessor.Protocol, accessor.Host, *accessor.Port, url)
    } else {
        return fmt.Sprintf("%s://%s/%s", accessor.Protocol, accessor.Host, url)
    }
}

// function to generate new HTTP request with JSON settings for
// a given method, url and body
func(accessor *BaseAPIAccessor) NewJSONRequest(method, url string, body io.Reader,
    headers map[string]string) (*http.Request, error) {
    // generate new HTTP request with given settings
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        log.Error(fmt.Errorf("unable to generate new HTTP Request: %+v", err))
        return nil, err
    }
    // set JSON as content type and return
    req.Header.Set("Content-Type", "application/json")
    // set additional headers provided
    for header, val := range(headers) {
        req.Header.Set(header, val)
    }
    return req, nil
}